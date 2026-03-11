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

// Package v1alpha5 provides a dynamic client for the VM Operator v1alpha5 API.
package v1alpha5

import (
	"context"

	vmopv5 "github.com/vmware-tanzu/vm-operator/api/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var (
	// VirtualMachineGVR is the GroupVersionResource for v1alpha5 VirtualMachine.
	VirtualMachineGVR = schema.GroupVersionResource{
		Group:    "vmoperator.vmware.com",
		Version:  "v1alpha5",
		Resource: "virtualmachines",
	}
	// VirtualMachineServiceGVR is the GroupVersionResource for v1alpha5 VirtualMachineService.
	VirtualMachineServiceGVR = schema.GroupVersionResource{
		Group:    "vmoperator.vmware.com",
		Version:  "v1alpha5",
		Resource: "virtualmachineservices",
	}
)

// Client wraps a dynamic client for the v1alpha5 API group.
type Client struct {
	dynamicClient dynamic.Interface
}

// NewForConfig creates a new v1alpha5 Client from the given REST config.
func NewForConfig(cfg *rest.Config) (*Client, error) {
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{dynamicClient: dc}, nil
}

// NewWithDynamicClient creates a v1alpha5 Client from an existing dynamic.Interface.
// This is intended for testing with a fake dynamic client.
func NewWithDynamicClient(dc dynamic.Interface) *Client {
	return &Client{dynamicClient: dc}
}

// GetVirtualMachine fetches a VirtualMachine by namespace and name.
func (c *Client) GetVirtualMachine(ctx context.Context, namespace, name string) (*vmopv5.VirtualMachine, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	vm := &vmopv5.VirtualMachine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), vm); err != nil {
		return nil, err
	}
	return vm, nil
}

// ListVirtualMachines lists VirtualMachines in the given namespace.
func (c *Client) ListVirtualMachines(ctx context.Context, namespace string, opts metav1.ListOptions) (*vmopv5.VirtualMachineList, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineGVR).Namespace(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	list := &vmopv5.VirtualMachineList{}
	for i := range obj.Items {
		vm := &vmopv5.VirtualMachine{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Items[i].UnstructuredContent(), vm); err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *vm)
	}
	return list, nil
}

// GetVirtualMachineService fetches a VirtualMachineService by namespace and name.
func (c *Client) GetVirtualMachineService(ctx context.Context, namespace, name string) (*vmopv5.VirtualMachineService, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	vms := &vmopv5.VirtualMachineService{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), vms); err != nil {
		return nil, err
	}
	return vms, nil
}

// ListVirtualMachineServices lists VirtualMachineServices in the given namespace.
func (c *Client) ListVirtualMachineServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*vmopv5.VirtualMachineServiceList, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	list := &vmopv5.VirtualMachineServiceList{}
	for i := range obj.Items {
		vms := &vmopv5.VirtualMachineService{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Items[i].UnstructuredContent(), vms); err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *vms)
	}
	return list, nil
}

// CreateVirtualMachineService creates a VirtualMachineService.
func (c *Client) CreateVirtualMachineService(ctx context.Context, vms *vmopv5.VirtualMachineService) (*vmopv5.VirtualMachineService, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vms)
	if err != nil {
		return nil, err
	}
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(vms.Namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	created := &vmopv5.VirtualMachineService{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), created); err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateVirtualMachineService updates a VirtualMachineService.
func (c *Client) UpdateVirtualMachineService(ctx context.Context, vms *vmopv5.VirtualMachineService) (*vmopv5.VirtualMachineService, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vms)
	if err != nil {
		return nil, err
	}
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(vms.Namespace).Update(ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	updated := &vmopv5.VirtualMachineService{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteVirtualMachineService deletes a VirtualMachineService by namespace and name.
func (c *Client) DeleteVirtualMachineService(ctx context.Context, namespace, name string) error {
	return c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
