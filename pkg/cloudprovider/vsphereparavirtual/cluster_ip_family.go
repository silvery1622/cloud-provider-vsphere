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

package vsphereparavirtual

import (
	"fmt"
	"strings"

	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator/factory"
)

// ClusterIPFamilyIPv4 is the default: IPv4-first NodeInternalIP ordering; does
// not require a VM Operator API with dual-stack VirtualMachineService fields.
const ClusterIPFamilyIPv4 = "ipv4"

// ClusterIPFamilyIPv6 is IPv6-first ordering and requires
// --vm-operator-api-version >= v1alpha6 (dual-stack VirtualMachineService fields).
const ClusterIPFamilyIPv6 = "ipv6"

// ClusterIPFamilyIPv4IPv6 is dual-stack with IPv4 before IPv6 in the reported
// NodeInternalIP list; requires --vm-operator-api-version >= v1alpha6.
const ClusterIPFamilyIPv4IPv6 = "ipv4ipv6"

// ClusterIPFamilyIPv6IPv4 is dual-stack with IPv6 before IPv4 (same primary
// order as ClusterIPFamilyIPv6); requires --vm-operator-api-version >= v1alpha6.
const ClusterIPFamilyIPv6IPv4 = "ipv6ipv4"

// vmopAPILevel assigns monotonic levels to supported --vm-operator-api-version
// values. It must stay in sync with factory.NewAdapter: any version accepted there
// that supports dual-stack VirtualMachineService fields needs level >= vmopAPILevelV1alpha6.
// When adding a new supported version, add a case to TestVmopSupportsDualStackVMServiceAPI_levels.
var vmopAPILevel = map[string]int{
	factory.V1alpha2: 20,
	factory.V1alpha5: 50,
	factory.V1alpha6: 60,
}

const vmopAPILevelV1alpha6 = 60

// ParseClusterIPFamily validates the --cluster-ip-family flag and returns a
// canonical lowercase value: ipv4, ipv6, ipv4ipv6, or ipv6ipv4.
func ParseClusterIPFamily(raw string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "ipv4":
		return ClusterIPFamilyIPv4, nil
	case "ipv6":
		return ClusterIPFamilyIPv6, nil
	case "ipv4ipv6":
		return ClusterIPFamilyIPv4IPv6, nil
	case "ipv6ipv4":
		return ClusterIPFamilyIPv6IPv4, nil
	default:
		return "", fmt.Errorf("invalid --cluster-ip-family %q: must be one of %s, %s, %s, %s",
			raw, ClusterIPFamilyIPv4, ClusterIPFamilyIPv6, ClusterIPFamilyIPv4IPv6, ClusterIPFamilyIPv6IPv4)
	}
}

// vmopSupportsDualStackVMServiceAPI reports whether vmopAPIVersion is known and
// carries VirtualMachineService dual-stack fields in practice (level >= v1alpha6).
// An unknown version returns false; add new versions to vmopAPILevel before relying on this function.
func vmopSupportsDualStackVMServiceAPI(version string) bool {
	level, ok := vmopAPILevel[version]
	return ok && level >= vmopAPILevelV1alpha6
}

// clusterIPFamilyRequiresDualStackVMOP reports whether the canonical family
// needs VM Operator APIs that persist dual-stack LoadBalancer fields.
func clusterIPFamilyRequiresDualStackVMOP(clusterIPFamily string) bool {
	switch clusterIPFamily {
	case ClusterIPFamilyIPv6, ClusterIPFamilyIPv4IPv6, ClusterIPFamilyIPv6IPv4:
		return true
	default:
		return false
	}
}

// validateIPFamilyConfig ensures ipv6 / ipv4ipv6 / ipv6ipv4 are only used with
// a VM Operator API version >= v1alpha6 (see vmopAPILevel). ipv4 does not
// require dual-stack VirtualMachineService fields.
//
// clusterIPFamily must be a canonical value returned from ParseClusterIPFamily.
func validateIPFamilyConfig(clusterIPFamily, vmopAPIVersion string) error {
	if !clusterIPFamilyRequiresDualStackVMOP(clusterIPFamily) {
		return nil
	}
	if !vmopSupportsDualStackVMServiceAPI(vmopAPIVersion) {
		return fmt.Errorf(
			"--cluster-ip-family=%s requires --vm-operator-api-version >= %s (dual-stack VirtualMachineService fields), got %q; "+
				"earlier API versions omit ipFamilies/ipFamilyPolicy and can silently provision IPv4-only load balancers",
			clusterIPFamily, factory.V1alpha6, vmopAPIVersion,
		)
	}
	return nil
}
