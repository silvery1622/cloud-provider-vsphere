module k8s.io/cloud-provider-vsphere

go 1.15

require (
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/vmware/govmomi v0.22.1
	github.com/vmware/vsphere-automation-sdk-go/lib v0.2.0
	github.com/vmware/vsphere-automation-sdk-go/runtime v0.2.0
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt v0.3.0
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	google.golang.org/grpc v1.27.1
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/cloud-provider v0.20.2
	k8s.io/component-base v0.20.2
	k8s.io/klog/v2 v2.4.0
)
