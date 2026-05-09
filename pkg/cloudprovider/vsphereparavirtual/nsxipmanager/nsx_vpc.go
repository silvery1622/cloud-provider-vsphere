/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nsxipmanager

import (
	"context"

	vpcapisv1 "github.com/vmware-tanzu/nsx-operator/pkg/apis/vpc/v1alpha1"
	nsxclients "github.com/vmware-tanzu/nsx-operator/pkg/client/clientset/versioned"
	nsxinformers "github.com/vmware-tanzu/nsx-operator/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/routemanager/helper"
)

const (
	allocationSize             = 256
	ipv6AllocationPrefixLength = 80
)

var _ NSXIPManager = &NSXVPCIPManager{}

// NSXVPCIPManager is an implementation of NSXIPManager for NSX-T VPC.
type NSXVPCIPManager struct {
	client          nsxclients.Interface
	informerFactory nsxinformers.SharedInformerFactory
	svNamespace     string
	ownerRef        *metav1.OwnerReference
	podIPPoolType   string
	ipv4Enabled     bool
	ipv6Enabled     bool
}

// crNameForFamily returns the CR name for a given node and IP family.
// IPv4 keeps the bare node name (no rename for existing clusters).
// IPv6 always carries the helper.SuffixIPv6 ("-ipv6") suffix.
func crNameForFamily(nodeName string, ipv4 bool) string {
	if ipv4 {
		return nodeName
	}
	return nodeName + helper.SuffixIPv6
}

// createIPAddressAllocation creates one IPAddressAllocation CR for the given CR name and
// IP family. For IPv4 it sets IPAddressType=IPV4, visibility from podIPPoolType, and
// allocationSize=256. For IPv6 it sets IPAddressType=IPV6 and allocationPrefixLength=80;
// visibility does not apply and is left unset.
func (m *NSXVPCIPManager) createIPAddressAllocation(ctx context.Context, name string, ipv4 bool) error {
	klog.V(4).Infof("Creating IPAddressAllocation %s/%s (ipv4=%v)", m.svNamespace, name, ipv4)

	spec := vpcapisv1.IPAddressAllocationSpec{}
	if ipv4 {
		spec.IPAddressType = vpcapisv1.IPAddressTypeIPV4
		spec.IPAddressBlockVisibility = convertToIPAddressVisibility(m.podIPPoolType)
		spec.AllocationSize = allocationSize
	} else {
		spec.IPAddressType = vpcapisv1.IPAddressTypeIPV6
		spec.AllocationPrefixLength = ipv6AllocationPrefixLength
	}

	ipAddressAllocation := &vpcapisv1.IPAddressAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.svNamespace,
			OwnerReferences: []metav1.OwnerReference{
				*m.ownerRef,
			},
		},
		Spec: spec,
	}
	_, err := m.client.CrdV1alpha1().IPAddressAllocations(m.svNamespace).Create(ctx, ipAddressAllocation, metav1.CreateOptions{})
	return err
}

// ClaimPodCIDR will claim pod cidr for the node by creating IPAddressAllocation CRs (one per enabled family).
func (m *NSXVPCIPManager) ClaimPodCIDR(node *corev1.Node) error {
	// Compute the number of families we need to provision.
	expectedCount := 0
	if m.ipv4Enabled {
		expectedCount++
	}
	if m.ipv6Enabled {
		expectedCount++
	}

	// Fast-path: skip CR creation if the node already has all expected family CIDRs.
	// This is best-effort; the createIPAddressAllocation calls below are idempotent
	// and will handle the case where a CR already exists (lister Get avoids redundant
	// API calls, and the API server rejects duplicate names with AlreadyExists).
	if node.Spec.PodCIDR != "" && len(node.Spec.PodCIDRs) >= expectedCount {
		klog.V(4).Infof("Node %s already has %d pod CIDR(s), skipping claim", node.Name, len(node.Spec.PodCIDRs))
		return nil
	}

	lister := m.informerFactory.Crd().V1alpha1().IPAddressAllocations().Lister().IPAddressAllocations(m.svNamespace)

	// context.TODO() is used here because the NSXIPManager interface does not yet
	// carry a context; a future refactor can thread the caller's context through.
	ctx := context.TODO()

	if m.ipv4Enabled {
		crName := crNameForFamily(node.Name, true)
		if _, err := lister.Get(crName); err != nil {
			if apierrors.IsNotFound(err) {
				if err := m.createIPAddressAllocation(ctx, crName, true); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			klog.V(4).Infof("Node %s IPv4 IPAddressAllocation %s already exists", node.Name, crName)
		}
	}

	if m.ipv6Enabled {
		crName := crNameForFamily(node.Name, false)
		if _, err := lister.Get(crName); err != nil {
			if apierrors.IsNotFound(err) {
				if err := m.createIPAddressAllocation(ctx, crName, false); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			klog.V(4).Infof("Node %s IPv6 IPAddressAllocation %s already exists", node.Name, crName)
		}
	}

	return nil
}

// ReleasePodCIDR unconditionally deletes all per-family IPAddressAllocation CRs for
// the node, ignoring NotFound errors. A missing CR is treated as already cleaned up.
// This guarantees that a partially realized dual stack cluster does not leak NSX state
// when it is torn down.
func (m *NSXVPCIPManager) ReleasePodCIDR(node *corev1.Node) error {
	// context.TODO() is used here because the NSXIPManager interface does not yet
	// carry a context; a future refactor can thread the caller's context through.
	ctx := context.TODO()
	if m.ipv4Enabled {
		if err := m.deleteIPAddressAllocation(ctx, node.Name, true); err != nil {
			return err
		}
	}
	if m.ipv6Enabled {
		if err := m.deleteIPAddressAllocation(ctx, node.Name, false); err != nil {
			return err
		}
	}
	return nil
}

// deleteIPAddressAllocation deletes the IPAddressAllocation CR for the given node and
// IP family. NotFound is treated as success (idempotent delete).
func (m *NSXVPCIPManager) deleteIPAddressAllocation(ctx context.Context, nodeName string, ipv4 bool) error {
	crName := crNameForFamily(nodeName, ipv4)
	klog.V(4).Infof("Deleting IPAddressAllocation %s/%s", m.svNamespace, crName)
	if err := m.client.CrdV1alpha1().IPAddressAllocations(m.svNamespace).Delete(ctx, crName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		klog.V(4).Infof("IPAddressAllocation %s/%s not found, treating as already deleted", m.svNamespace, crName)
	}
	return nil
}

// convertToIPAddressVisibility converts the ip pool type to the ip address visibility. This is needed because the nsx
// does not unify names yet. Public equals to External.
func convertToIPAddressVisibility(ipPoolType string) vpcapisv1.IPAddressVisibility {
	if ipPoolType == PublicIPPoolType {
		return vpcapisv1.IPAddressVisibilityExternal
	}
	return vpcapisv1.IPAddressVisibilityPrivate
}

// NewNSXVPCIPManager returns an NSXIPManager that manages IPAddressAllocation CRs in VPC mode.
// ipv4Enabled and ipv6Enabled control which per-family CRs are created or deleted.
func NewNSXVPCIPManager(client nsxclients.Interface, informerFactory nsxinformers.SharedInformerFactory, svNamespace, podIPPoolType string, ownerRef *metav1.OwnerReference, ipv4Enabled, ipv6Enabled bool) NSXIPManager {
	return &NSXVPCIPManager{
		client:          client,
		informerFactory: informerFactory,
		svNamespace:     svNamespace,
		ownerRef:        ownerRef,
		podIPPoolType:   podIPPoolType,
		ipv4Enabled:     ipv4Enabled,
		ipv6Enabled:     ipv6Enabled,
	}
}
