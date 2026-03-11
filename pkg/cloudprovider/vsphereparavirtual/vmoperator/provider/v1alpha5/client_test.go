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

package v1alpha5

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	vmopv5 "github.com/vmware-tanzu/vm-operator/api/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	dynamicfake "k8s.io/client-go/dynamic/fake"
)

const testNS = "test-ns"

func newTestClient() (*Client, *dynamicfake.FakeDynamicClient) {
	scheme := runtime.NewScheme()
	_ = vmopv5.AddToScheme(scheme)
	fc := dynamicfake.NewSimpleDynamicClient(scheme)
	return NewWithDynamicClient(fc), fc
}

// seedVM seeds a VirtualMachine into the fake dynamic client directly.
func seedVM(t *testing.T, fc *dynamicfake.FakeDynamicClient, vm *vmopv5.VirtualMachine) {
	t.Helper()
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vm)
	assert.NoError(t, err)
	_, err = fc.Resource(VirtualMachineGVR).Namespace(vm.Namespace).Create(
		context.Background(), &unstructured.Unstructured{Object: obj}, metav1.CreateOptions{})
	assert.NoError(t, err)
}

func TestGetVirtualMachine(t *testing.T) {
	testCases := []struct {
		name       string
		seedName   string
		queryName  string
		getFunc    func(clientgotesting.Action) (bool, runtime.Object, error)
		expectBios string
		expectErr  bool
	}{
		{
			name:       "returns VM when it exists",
			seedName:   "vm-1",
			queryName:  "vm-1",
			expectBios: "bios-1",
		},
		{
			name:      "returns error when VM does not exist",
			seedName:  "other-vm",
			queryName: "nonexistent",
			expectErr: true,
		},
		{
			name:      "returns error on API failure",
			seedName:  "vm-1",
			queryName: "vm-1",
			getFunc: func(clientgotesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("api error")
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, fc := newTestClient()
			seedVM(t, fc, &vmopv5.VirtualMachine{
				ObjectMeta: metav1.ObjectMeta{Name: tc.seedName, Namespace: testNS},
				Status:     vmopv5.VirtualMachineStatus{BiosUUID: "bios-1"},
			})
			if tc.getFunc != nil {
				fc.PrependReactor("get", "virtualmachines", tc.getFunc)
			}

			vm, err := c.GetVirtualMachine(context.Background(), testNS, tc.queryName)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, vm)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectBios, vm.Status.BiosUUID)
		})
	}
}

func TestListVirtualMachines(t *testing.T) {
	testCases := []struct {
		name        string
		seedVMs     []string
		listFunc    func(clientgotesting.Action) (bool, runtime.Object, error)
		expectedLen int
		expectErr   bool
	}{
		{
			name:        "returns empty list when no VMs exist",
			seedVMs:     nil,
			expectedLen: 0,
		},
		{
			name:        "returns all VMs in namespace",
			seedVMs:     []string{"vm-1", "vm-2", "vm-3"},
			expectedLen: 3,
		},
		{
			name:    "returns error on API failure",
			seedVMs: []string{"vm-1"},
			listFunc: func(clientgotesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("api error")
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, fc := newTestClient()
			for _, name := range tc.seedVMs {
				seedVM(t, fc, &vmopv5.VirtualMachine{
					ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNS},
				})
			}
			if tc.listFunc != nil {
				fc.PrependReactor("list", "virtualmachines", tc.listFunc)
			}

			list, err := c.ListVirtualMachines(context.Background(), testNS, metav1.ListOptions{})
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, list.Items, tc.expectedLen)
		})
	}
}

func TestVirtualMachineServiceCRUD(t *testing.T) {
	testCases := []struct {
		name       string
		vmsName    string
		createFunc func(clientgotesting.Action) (bool, runtime.Object, error)
		expectErr  bool
	}{
		{
			name:    "create, get, update, list, delete succeeds",
			vmsName: "test-vms",
		},
		{
			name:    "create returns error on API failure",
			vmsName: "test-vms",
			createFunc: func(clientgotesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("api error")
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, fc := newTestClient()
			if tc.createFunc != nil {
				fc.PrependReactor("create", "virtualmachineservices", tc.createFunc)
			}

			vms := &vmopv5.VirtualMachineService{
				ObjectMeta: metav1.ObjectMeta{Name: tc.vmsName, Namespace: testNS},
				Spec: vmopv5.VirtualMachineServiceSpec{
					Type: vmopv5.VirtualMachineServiceTypeLoadBalancer,
				},
			}

			created, err := c.CreateVirtualMachineService(context.Background(), vms)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.vmsName, created.Name)

			// Get
			got, err := c.GetVirtualMachineService(context.Background(), testNS, tc.vmsName)
			assert.NoError(t, err)
			assert.Equal(t, tc.vmsName, got.Name)

			// Update
			got.Spec.LoadBalancerIP = "1.2.3.4"
			updated, err := c.UpdateVirtualMachineService(context.Background(), got)
			assert.NoError(t, err)
			assert.Equal(t, "1.2.3.4", updated.Spec.LoadBalancerIP)

			// List
			list, err := c.ListVirtualMachineServices(context.Background(), testNS, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Len(t, list.Items, 1)

			// Delete
			err = c.DeleteVirtualMachineService(context.Background(), testNS, tc.vmsName)
			assert.NoError(t, err)
		})
	}
}
