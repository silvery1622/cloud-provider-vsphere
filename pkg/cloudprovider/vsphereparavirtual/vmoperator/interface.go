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

// Package vmoperator provides version-agnostic interfaces for interacting with
// VM Operator resources. Business logic in the CPI should depend only on these
// interfaces and the hub types in the types sub-package, never on a specific
// VM Operator API version.
package vmoperator

import (
	"context"

	"k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator/types"
)

// Interface is the main version-agnostic interface for VM Operator operations.
// Implementations are provided by version-specific adapters selected at startup
// via the --vm-operator-api-version flag. All CPI business logic must depend only on
// this interface; it must never import a versioned VM Operator API package directly.
type Interface interface {
	// VirtualMachines returns an interface for VirtualMachine operations.
	VirtualMachines() VirtualMachineInterface
	// VirtualMachineServices returns an interface for VirtualMachineService operations.
	VirtualMachineServices() VirtualMachineServiceInterface
}

// VirtualMachineInterface has methods to work with VirtualMachine resources
// using version-agnostic hub types.
type VirtualMachineInterface interface {
	// Get returns the VirtualMachine with the given name in the given namespace.
	Get(ctx context.Context, namespace, name string) (*types.VirtualMachineInfo, error)
	// List returns all VirtualMachines in the given namespace matching opts.
	List(ctx context.Context, namespace string, opts types.ListOptions) ([]*types.VirtualMachineInfo, error)
	// GetByBiosUUID returns the VirtualMachine whose BiosUUID matches the given
	// uuid, or nil if no match is found. BiosUUID is a status field and cannot
	// be filtered server-side via FieldSelector; implementations must List all
	// VMs in the namespace and scan in memory. Callers should prefer Get by name
	// when the VM name is available to avoid this O(n) scan.
	GetByBiosUUID(ctx context.Context, namespace, biosUUID string) (*types.VirtualMachineInfo, error)
}

// VirtualMachineServiceInterface has methods to work with VirtualMachineService
// resources using version-agnostic hub types.
type VirtualMachineServiceInterface interface {
	// Get returns the VirtualMachineService with the given name in the given namespace.
	Get(ctx context.Context, namespace, name string) (*types.VirtualMachineServiceInfo, error)
	// List returns all VirtualMachineServices in the given namespace matching opts.
	List(ctx context.Context, namespace string, opts types.ListOptions) ([]*types.VirtualMachineServiceInfo, error)
	// Create creates a new VirtualMachineService from the given hub info.
	Create(ctx context.Context, vms *types.VirtualMachineServiceInfo) (*types.VirtualMachineServiceInfo, error)
	// Update replaces the mutable fields of an existing VirtualMachineService identified
	// by namespace/name with the fields provided in update. This is a non-atomic
	// read-modify-write operation; callers should handle 409 Conflict by retrying.
	Update(ctx context.Context, namespace, name string, update *types.VirtualMachineServiceInfo) (*types.VirtualMachineServiceInfo, error)
	// Delete deletes the VirtualMachineService with the given name in the given namespace.
	Delete(ctx context.Context, namespace, name string) error
}
