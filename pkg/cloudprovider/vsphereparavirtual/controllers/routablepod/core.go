/*
Copyright 2021 The Kubernetes Authors.
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

package routablepod

import (
	"context"
	"fmt"

	nsxclients "github.com/vmware-tanzu/nsx-operator/pkg/client/clientset/versioned"
	nsxinformers "github.com/vmware-tanzu/nsx-operator/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/controllers/routablepod/ipaddressallocation"
	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/controllers/routablepod/ippool"
	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/controllers/routablepod/node"
	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/ippoolmanager/helper"
	ippmv1alpha1 "k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/ippoolmanager/v1alpha1"
	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/nsxipmanager"
	k8s "k8s.io/cloud-provider-vsphere/pkg/common/kubernetes"
)

// StartControllers starts the Routable Pods controllers: in VPC mode it starts
// ipaddressallocation_controller (which patches node PodCIDRs) and node_controller
// (which manages IPAddressAllocation CRs); in T1 mode it starts ippool_controller
// and node_controller.
func StartControllers(scCfg *rest.Config, client kubernetes.Interface,
	informerManager *k8s.InformerManager, clusterName, clusterNS string, ownerRef *metav1.OwnerReference,
	vpcModeEnabled bool, podIPPoolType string,
	ipv4Enabled, ipv6Enabled, primaryIPFamilyIsIPv4 bool) error {

	if clusterName == "" {
		return fmt.Errorf("cluster name can't be empty")
	}
	if clusterNS == "" {
		return fmt.Errorf("cluster namespace can't be empty")
	}

	klog.V(2).Info("Routable pod controllers start with VPC mode enabled: ", vpcModeEnabled)

	ctx := informerManager.GetContext()

	// Derive expectedFamilyCount from the enabled families directly so it is
	// always consistent with what NSXVPCIPManager sees.
	expectedFamilyCount := 0
	if ipv4Enabled {
		expectedFamilyCount++
	}
	if ipv6Enabled {
		expectedFamilyCount++
	}
	// dualStackEnabled is derived locally from the enabled families rather than
	// being accepted as a parameter, keeping StartControllers' API minimal and
	// ensuring it stays consistent with ipv4Enabled/ipv6Enabled.
	dualStackEnabled := ipv4Enabled && ipv6Enabled

	var nsxIPManager nsxipmanager.NSXIPManager
	if vpcModeEnabled {
		nsxClient, nsxInformerFactory, err := getNSXClientAndInformer(scCfg, clusterNS)
		if err != nil {
			return fmt.Errorf("fail to get NSX client or informer factory: %w", err)
		}

		startIPAddressAllocationController(ctx, client, informerManager, nsxInformerFactory, dualStackEnabled, primaryIPFamilyIsIPv4)

		nsxIPManager = nsxipmanager.NewNSXVPCIPManager(nsxClient, nsxInformerFactory, clusterNS, podIPPoolType, ownerRef, ipv4Enabled, ipv6Enabled)
	} else {
		ippManager, err := ippmv1alpha1.NewIPPoolManager(scCfg, clusterNS)
		if err != nil {
			return fmt.Errorf("fail to get ippool manager or start ippool controller: %w", err)
		}

		startIPPoolController(ctx, client, ippManager)

		nsxIPManager = nsxipmanager.NewNSXT1IPManager(ippManager, clusterName, clusterNS, ownerRef)
	}

	nodeController := node.NewController(client, nsxIPManager, informerManager, clusterName, clusterNS, ownerRef, expectedFamilyCount)
	go nodeController.Run(ctx.Done())

	return nil
}

func startIPAddressAllocationController(ctx context.Context, client kubernetes.Interface, informerManager *k8s.InformerManager, nsxInformerFactory nsxinformers.SharedInformerFactory, dualStackEnabled, primaryIPFamilyIsIPv4 bool) {
	ipAddressAllocationController := ipaddressallocation.NewController(
		ctx,
		client,
		informerManager.GetNodeLister(),
		informerManager.IsNodeInformerSynced(),
		nsxInformerFactory.Crd().V1alpha1().IPAddressAllocations(),
		dualStackEnabled,
		primaryIPFamilyIsIPv4)
	go ipAddressAllocationController.Run(ctx, 1)
	nsxInformerFactory.Start(ctx.Done())
}

func startIPPoolController(ctx context.Context, client kubernetes.Interface, ippManager *ippmv1alpha1.IPPoolManager) {
	ippoolController := ippool.NewController(client, ippManager)
	go ippoolController.Run(ctx.Done())
	ippManager.StartIPPoolInformers(ctx.Done())
}

func getNSXClientAndInformer(svCfg *rest.Config, svNamespace string) (nsxclients.Interface, nsxinformers.SharedInformerFactory, error) {
	client, err := nsxclients.NewForConfig(svCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error building nsx-operator clientset: %w", err)
	}

	informerFactory := nsxinformers.NewSharedInformerFactoryWithOptions(client,
		helper.DefaultResyncTime, nsxinformers.WithNamespace(svNamespace))

	return client, informerFactory, nil
}
