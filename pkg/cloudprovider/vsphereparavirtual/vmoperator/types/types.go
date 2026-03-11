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

// Package types defines version-agnostic hub types for VM Operator resources.
// These types are a minimal union of all fields needed by the CPI across all
// supported VM Operator API versions. Fields are never removed or renamed so
// that older version adapters continue to function correctly.
package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PowerState is the hub type for a VM's power state.
// Values are intentionally kept stable across all supported API versions.
type PowerState string

const (
	// PowerStatePoweredOn indicates the VM is powered on.
	PowerStatePoweredOn PowerState = "PoweredOn"
	// PowerStatePoweredOff indicates the VM is powered off.
	PowerStatePoweredOff PowerState = "PoweredOff"
	// PowerStateSuspended indicates the VM is suspended.
	PowerStateSuspended PowerState = "Suspended"
)

// VirtualMachineServiceType is the hub type for a VirtualMachineService type.
type VirtualMachineServiceType string

const (
	// VirtualMachineServiceTypeLoadBalancer is the LoadBalancer service type.
	VirtualMachineServiceTypeLoadBalancer VirtualMachineServiceType = "LoadBalancer"
	// VirtualMachineServiceTypeClusterIP is the ClusterIP service type.
	VirtualMachineServiceTypeClusterIP VirtualMachineServiceType = "ClusterIP"
	// VirtualMachineServiceTypeExternalName is the ExternalName service type.
	VirtualMachineServiceTypeExternalName VirtualMachineServiceType = "ExternalName"
)

// VirtualMachineInfo is the hub type for a VirtualMachine resource.
// It contains only the fields required by the CPI for node discovery.
type VirtualMachineInfo struct {
	Name       string
	Namespace  string
	Labels     map[string]string
	BiosUUID   string
	PowerState PowerState
	// PrimaryIP4 is the primary IPv4 address of the VM.
	PrimaryIP4 string
	// PrimaryIP6 is the primary IPv6 address of the VM.
	PrimaryIP6 string
}

// VirtualMachineServicePort describes a single port exposed by a VirtualMachineService.
type VirtualMachineServicePort struct {
	Name string
	// Protocol must be one of "TCP", "UDP", or "SCTP", matching the values
	// accepted by the VM Operator API. It is kept as a plain string (rather
	// than a typed constant) because the CPI passes through whatever the
	// Kubernetes Service specifies without interpreting the value.
	Protocol   string
	Port       int32
	TargetPort int32
}

// VirtualMachineServiceSpec is the hub spec for a VirtualMachineService.
type VirtualMachineServiceSpec struct {
	Type                     VirtualMachineServiceType
	Ports                    []VirtualMachineServicePort
	Selector                 map[string]string
	LoadBalancerIP           string
	LoadBalancerSourceRanges []string
}

// LoadBalancerIngress represents a single ingress point for a load balancer.
type LoadBalancerIngress struct {
	IP       string
	Hostname string
}

// VirtualMachineServiceStatus is the hub status for a VirtualMachineService.
type VirtualMachineServiceStatus struct {
	LoadBalancerIngress []LoadBalancerIngress
}

// VirtualMachineServiceInfo is the hub type for a VirtualMachineService resource.
type VirtualMachineServiceInfo struct {
	Name            string
	Namespace       string
	Labels          map[string]string
	Annotations     map[string]string
	OwnerReferences []metav1.OwnerReference
	Spec            VirtualMachineServiceSpec
	Status          VirtualMachineServiceStatus
}

// ListOptions contains options for listing resources.
type ListOptions struct {
	LabelSelector string
}
