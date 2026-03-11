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

package client

import (
	"context"

	vmopv1 "github.com/vmware-tanzu/vm-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetVirtualMachine fetches a VirtualMachine by namespace and name.
func (c *Client) GetVirtualMachine(ctx context.Context, namespace, name string) (*vmopv1.VirtualMachine, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	vm := &vmopv1.VirtualMachine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), vm); err != nil {
		return nil, err
	}
	return vm, nil
}

// ListVirtualMachines lists VirtualMachines in the given namespace.
func (c *Client) ListVirtualMachines(ctx context.Context, namespace string, opts metav1.ListOptions) (*vmopv1.VirtualMachineList, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineGVR).Namespace(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	list := &vmopv1.VirtualMachineList{}
	for i := range obj.Items {
		vm := &vmopv1.VirtualMachine{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Items[i].UnstructuredContent(), vm); err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *vm)
	}
	return list, nil
}

// CreateVirtualMachine creates a VirtualMachine.
func (c *Client) CreateVirtualMachine(ctx context.Context, vm *vmopv1.VirtualMachine) (*vmopv1.VirtualMachine, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vm)
	if err != nil {
		return nil, err
	}
	obj, err := c.dynamicClient.Resource(VirtualMachineGVR).Namespace(vm.Namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	created := &vmopv1.VirtualMachine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), created); err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateVirtualMachine updates a VirtualMachine.
func (c *Client) UpdateVirtualMachine(ctx context.Context, vm *vmopv1.VirtualMachine) (*vmopv1.VirtualMachine, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vm)
	if err != nil {
		return nil, err
	}
	obj, err := c.dynamicClient.Resource(VirtualMachineGVR).Namespace(vm.Namespace).Update(ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	updated := &vmopv1.VirtualMachine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteVirtualMachine deletes a VirtualMachine by namespace and name.
func (c *Client) DeleteVirtualMachine(ctx context.Context, namespace, name string) error {
	return c.dynamicClient.Resource(VirtualMachineGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
