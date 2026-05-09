/* Copyright © 2024 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: Apache-2.0 */

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type IPAddressVisibility string

var (
	IPAddressVisibilityExternal   IPAddressVisibility = "External"
	IPAddressVisibilityPrivate    IPAddressVisibility = "Private"
	IPAddressVisibilityPrivateTGW IPAddressVisibility = "PrivateTGW"
)

// IPAddressTypeIPV4 and IPAddressTypeIPV6 are the allowed values for IPAddressAllocationSpec.IPAddressType.
// The IPAddressType base type is declared in subnet_types.go for the whole v1alpha1 package.
const (
	IPAddressTypeIPV4 IPAddressType = "IPV4"
	IPAddressTypeIPV6 IPAddressType = "IPV6"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// IPAddressAllocation is the Schema for the IP allocation API.
// +kubebuilder:printcolumn:name="IPAddressBlockVisibility",type=string,JSONPath=`.spec.ipAddressBlockVisibility`,description="IPAddressBlockVisibility of IPAddressAllocation"
// +kubebuilder:printcolumn:name="AllocationIPs",type=string,JSONPath=`.status.allocationIPs`, description="AllocationIPs for the IPAddressAllocation"
type IPAddressAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   IPAddressAllocationSpec   `json:"spec"`
	Status IPAddressAllocationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPAddressAllocationList contains a list of IPAddressAllocation.
type IPAddressAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAddressAllocation `json:"items"`
}

// IPAddressAllocationSpec defines the desired state of IPAddressAllocation.
// +kubebuilder:validation:XValidation:rule="!has(self.allocationSize) || !has(self.allocationIPs)", message="Only one of allocationSize or allocationIPs can be specified"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.allocationSize) || has(self.allocationSize)", message="allocationSize is required once set"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.allocationIPs) || has(self.allocationIPs)", message="allocationIPs is required once set"
type IPAddressAllocationSpec struct {
	// IPAddressBlockVisibility specifies the visibility of the IPBlocks to allocate IP addresses. Can be External, Private or PrivateTGW.
	// Not applicable when IPAddressType is IPV6.
	// +kubebuilder:validation:Enum=External;Private;PrivateTGW
	// +kubebuilder:default=Private
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	IPAddressBlockVisibility IPAddressVisibility `json:"ipAddressBlockVisibility,omitempty"`
	// IPAddressType specifies the IP address family. Defaults to IPV4.
	// +kubebuilder:default=IPV4
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	IPAddressType IPAddressType `json:"ipAddressType,omitempty"`
	// AllocationSize specifies the size of allocationIPs to be allocated.
	// It should be a power of 2. Not applicable when IPAddressType is IPV6.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:Minimum:=1
	AllocationSize int `json:"allocationSize,omitempty"`
	// AllocationPrefixLength specifies the prefix length for IPv6 allocation.
	// Only applicable when IPAddressType is IPV6. Valid range: 64-128.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:Minimum:=64
	// +kubebuilder:validation:Maximum:=128
	// +optional
	AllocationPrefixLength int `json:"allocationPrefixLength,omitempty"`
	// AllocationIPs specifies the Allocated IP addresses in CIDR or single IP Address format.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	AllocationIPs string `json:"allocationIPs,omitempty"`
}

// IPAddressAllocationStatus defines the observed state of IPAddressAllocation.
type IPAddressAllocationStatus struct {
	// AllocationIPs is the allocated IP addresses
	AllocationIPs string      `json:"allocationIPs"`
	Conditions    []Condition `json:"conditions,omitempty"`
}

func init() {
	SchemeBuilder.Register(&IPAddressAllocation{}, &IPAddressAllocationList{})
}
