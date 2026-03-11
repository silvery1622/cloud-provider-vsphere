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

package v1alpha2_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	vmopv1 "github.com/vmware-tanzu/vm-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	adapterv2 "k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator/adapter/v1alpha2"
	vmopclient "k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator/client"
	vmoptypes "k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator/types"
)

const testNS = "test-ns"

func TestAdapter_VirtualMachines_Get(t *testing.T) {
	testCases := []struct {
		name          string
		seedVM        *vmopv1.VirtualMachine
		queryName     string
		expectedBios  string
		expectedIP4   string
		expectedIP6   string
		expectedPower vmoptypes.PowerState
		expectedLabel map[string]string
		expectErr     bool
	}{
		{
			name: "returns full VM info when VM exists",
			seedVM: &vmopv1.VirtualMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vm",
					Namespace: testNS,
					Labels:    map[string]string{"zone": "us-east-1"},
				},
				Status: vmopv1.VirtualMachineStatus{
					BiosUUID:   "test-uuid",
					PowerState: vmopv1.VirtualMachinePowerStateOn,
					Network: &vmopv1.VirtualMachineNetworkStatus{
						PrimaryIP4: "10.0.0.1",
						PrimaryIP6: "fd00::1",
					},
				},
			},
			queryName:     "test-vm",
			expectedBios:  "test-uuid",
			expectedIP4:   "10.0.0.1",
			expectedIP6:   "fd00::1",
			expectedPower: vmoptypes.PowerStatePoweredOn,
			expectedLabel: map[string]string{"zone": "us-east-1"},
		},
		{
			name:      "returns error when VM does not exist",
			seedVM:    &vmopv1.VirtualMachine{ObjectMeta: metav1.ObjectMeta{Name: "other-vm", Namespace: testNS}},
			queryName: "nonexistent",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, fc := adapterv2.NewFakeAdapter()
			// Use a second Client wrapping the same fake to seed test data
			// independently of the adapter under test.
			c := vmopclient.NewFakeClient(fc)
			_, err := c.CreateVirtualMachine(context.Background(), tc.seedVM)
			assert.NoError(t, err)

			info, err := adapter.VirtualMachines().Get(context.Background(), testNS, tc.queryName)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, info)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.queryName, info.Name)
			assert.Equal(t, testNS, info.Namespace)
			assert.Equal(t, tc.expectedBios, info.BiosUUID)
			assert.Equal(t, tc.expectedPower, info.PowerState)
			assert.Equal(t, tc.expectedIP4, info.PrimaryIP4)
			assert.Equal(t, tc.expectedIP6, info.PrimaryIP6)
			assert.Equal(t, tc.expectedLabel, info.Labels)
		})
	}
}

func TestAdapter_VirtualMachines_List(t *testing.T) {
	testCases := []struct {
		name        string
		seedNames   []string
		expectedLen int
	}{
		{
			name:        "returns empty list when no VMs exist",
			seedNames:   nil,
			expectedLen: 0,
		},
		{
			name:        "returns all VMs in namespace",
			seedNames:   []string{"vm-1", "vm-2"},
			expectedLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, fc := adapterv2.NewFakeAdapter()
			// Use a second Client wrapping the same fake to seed test data.
			c := vmopclient.NewFakeClient(fc)
			for _, name := range tc.seedNames {
				_, err := c.CreateVirtualMachine(context.Background(), &vmopv1.VirtualMachine{
					ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNS},
				})
				assert.NoError(t, err)
			}

			list, err := adapter.VirtualMachines().List(context.Background(), testNS, vmoptypes.ListOptions{})
			assert.NoError(t, err)
			assert.Len(t, list, tc.expectedLen)
		})
	}
}

func TestAdapter_VirtualMachines_GetByBiosUUID(t *testing.T) {
	testCases := []struct {
		name         string
		seedVMs      []*vmopv1.VirtualMachine
		queryUUID    string
		expectedName string
		expectNil    bool
	}{
		{
			name: "returns VM info when BiosUUID matches",
			seedVMs: []*vmopv1.VirtualMachine{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "vm-a", Namespace: testNS},
					Status:     vmopv1.VirtualMachineStatus{BiosUUID: "uuid-a"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "vm-b", Namespace: testNS},
					Status:     vmopv1.VirtualMachineStatus{BiosUUID: "uuid-b"},
				},
			},
			queryUUID:    "uuid-b",
			expectedName: "vm-b",
		},
		{
			name: "returns nil when no VM matches BiosUUID",
			seedVMs: []*vmopv1.VirtualMachine{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "vm-a", Namespace: testNS},
					Status:     vmopv1.VirtualMachineStatus{BiosUUID: "uuid-a"},
				},
			},
			queryUUID: "nonexistent-uuid",
			expectNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, fc := adapterv2.NewFakeAdapter()
			// Use a second Client wrapping the same fake to seed test data.
			c := vmopclient.NewFakeClient(fc)
			for _, vm := range tc.seedVMs {
				_, err := c.CreateVirtualMachine(context.Background(), vm)
				assert.NoError(t, err)
			}

			info, err := adapter.VirtualMachines().GetByBiosUUID(context.Background(), testNS, tc.queryUUID)
			assert.NoError(t, err)
			if tc.expectNil {
				assert.Nil(t, info)
				return
			}
			assert.NotNil(t, info)
			assert.Equal(t, tc.expectedName, info.Name)
			assert.Equal(t, tc.queryUUID, info.BiosUUID)
		})
	}
}

func TestAdapter_VirtualMachineServices_CRUD(t *testing.T) {
	testCases := []struct {
		name        string
		createInfo  *vmoptypes.VirtualMachineServiceInfo
		updatePorts []vmoptypes.VirtualMachineServicePort
	}{
		{
			name: "full CRUD lifecycle succeeds",
			createInfo: &vmoptypes.VirtualMachineServiceInfo{
				Name:      "test-vms",
				Namespace: testNS,
				Spec: vmoptypes.VirtualMachineServiceSpec{
					Type: vmoptypes.VirtualMachineServiceTypeLoadBalancer,
					Ports: []vmoptypes.VirtualMachineServicePort{
						{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 30800},
					},
				},
			},
			updatePorts: []vmoptypes.VirtualMachineServicePort{
				{Name: "https", Protocol: "TCP", Port: 443, TargetPort: 30443},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, _ := adapterv2.NewFakeAdapter()

			// Create
			created, err := adapter.VirtualMachineServices().Create(context.Background(), tc.createInfo)
			assert.NoError(t, err)
			assert.Equal(t, tc.createInfo.Name, created.Name)
			assert.Len(t, created.Spec.Ports, len(tc.createInfo.Spec.Ports))

			// Get
			got, err := adapter.VirtualMachineServices().Get(context.Background(), testNS, tc.createInfo.Name)
			assert.NoError(t, err)
			assert.Equal(t, tc.createInfo.Name, got.Name)

			// Update
			update := &vmoptypes.VirtualMachineServiceInfo{
				Spec: vmoptypes.VirtualMachineServiceSpec{Ports: tc.updatePorts},
			}
			updated, err := adapter.VirtualMachineServices().Update(context.Background(), testNS, tc.createInfo.Name, update)
			assert.NoError(t, err)
			assert.Len(t, updated.Spec.Ports, 1)
			assert.Equal(t, tc.updatePorts[0].Name, updated.Spec.Ports[0].Name)

			// List
			list, err := adapter.VirtualMachineServices().List(context.Background(), testNS, vmoptypes.ListOptions{})
			assert.NoError(t, err)
			assert.Len(t, list, 1)

			// Delete
			err = adapter.VirtualMachineServices().Delete(context.Background(), testNS, tc.createInfo.Name)
			assert.NoError(t, err)

			// Verify deleted
			got, err = adapter.VirtualMachineServices().Get(context.Background(), testNS, tc.createInfo.Name)
			assert.Error(t, err)
			assert.Nil(t, got)
		})
	}
}
