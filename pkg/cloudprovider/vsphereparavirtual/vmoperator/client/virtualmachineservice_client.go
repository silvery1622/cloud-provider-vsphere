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

// GetVirtualMachineService fetches a VirtualMachineService by namespace and name.
func (c *Client) GetVirtualMachineService(ctx context.Context, namespace, name string) (*vmopv1.VirtualMachineService, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	vms := &vmopv1.VirtualMachineService{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), vms); err != nil {
		return nil, err
	}
	return vms, nil
}

// ListVirtualMachineServices lists VirtualMachineServices in the given namespace.
func (c *Client) ListVirtualMachineServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*vmopv1.VirtualMachineServiceList, error) {
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	list := &vmopv1.VirtualMachineServiceList{}
	for i := range obj.Items {
		vms := &vmopv1.VirtualMachineService{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Items[i].UnstructuredContent(), vms); err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *vms)
	}
	return list, nil
}

// CreateVirtualMachineService creates a VirtualMachineService.
func (c *Client) CreateVirtualMachineService(ctx context.Context, vms *vmopv1.VirtualMachineService) (*vmopv1.VirtualMachineService, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vms)
	if err != nil {
		return nil, err
	}
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(vms.Namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	created := &vmopv1.VirtualMachineService{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), created); err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateVirtualMachineService updates a VirtualMachineService.
func (c *Client) UpdateVirtualMachineService(ctx context.Context, vms *vmopv1.VirtualMachineService) (*vmopv1.VirtualMachineService, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vms)
	if err != nil {
		return nil, err
	}
	obj, err := c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(vms.Namespace).Update(ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	updated := &vmopv1.VirtualMachineService{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteVirtualMachineService deletes a VirtualMachineService by namespace and name.
func (c *Client) DeleteVirtualMachineService(ctx context.Context, namespace, name string) error {
	return c.dynamicClient.Resource(VirtualMachineServiceGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
