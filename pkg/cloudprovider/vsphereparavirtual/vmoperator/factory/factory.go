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

// Package factory provides a factory function that creates the correct
// vmoperator.Interface adapter based on the requested API version string.
// The version is typically supplied via the --vm-operator-api-version command-line
// flag, mirroring the approach used by other VKS components (e.g. CAPV).
package factory

import (
	"fmt"

	"k8s.io/client-go/rest"

	vmop "k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator"
	adapterv2 "k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator/adapter/v1alpha2"
	adapterv5 "k8s.io/cloud-provider-vsphere/pkg/cloudprovider/vsphereparavirtual/vmoperator/adapter/v1alpha5"
)

const (
	// V1alpha2 is the version string for the VM Operator v1alpha2 API.
	V1alpha2 = "v1alpha2"
	// V1alpha5 is the version string for the VM Operator v1alpha5 API.
	V1alpha5 = "v1alpha5"
)

// NewAdapter creates a vmoperator.Interface for the given API version using
// the provided REST config. Supported versions: v1alpha2, v1alpha5.
func NewAdapter(version string, cfg *rest.Config) (vmop.Interface, error) {
	switch version {
	case V1alpha2:
		return adapterv2.New(cfg)
	case V1alpha5:
		return adapterv5.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported vm-operator-api-version %q: supported versions are %s, %s",
			version, V1alpha2, V1alpha5)
	}
}
