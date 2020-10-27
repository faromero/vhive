module github.com/ustiugov/fccd-orchestrator

go 1.13

replace github.com/firecracker-microvm/firecracker-containerd => github.com/ustiugov/firecracker-containerd v0.0.0-20200804113524-bc259c9e8152

replace github.com/firecracker-microvm/firecracker-go-sdk => github.com/ustiugov/firecracker-go-sdk v0.20.1-0.20200625102438-8edf287b0123

require (
	github.com/containerd/containerd v1.3.6
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/ustiugov/fccd-orchestrator/ctriface v0.0.0-20201029105850-29220ab0dbd1
	github.com/ustiugov/fccd-orchestrator/helloworld v0.0.0-20200803195925-0629e1cf4599
	github.com/ustiugov/fccd-orchestrator/metrics v0.0.0-20200907081336-fae0d2f696c4
	github.com/ustiugov/fccd-orchestrator/proto v0.0.0-20200803195925-0629e1cf4599
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	google.golang.org/grpc v1.31.0
	k8s.io/cri-api v0.16.16-rc.0
)

// Workaround for github.com/containerd/containerd issue #3031
replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible

replace k8s.io/kubernetes => k8s.io/kubernetes v1.16.6

replace k8s.io/api => k8s.io/api v0.16.6

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.6

replace k8s.io/apimachinery => k8s.io/apimachinery v0.16.7-beta.0

replace k8s.io/apiserver => k8s.io/apiserver v0.16.6

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.6

replace k8s.io/client-go => k8s.io/client-go v0.16.6

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.6

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.6

replace k8s.io/code-generator => k8s.io/code-generator v0.16.7-beta.0

replace k8s.io/component-base => k8s.io/component-base v0.16.6

replace k8s.io/cri-api => k8s.io/cri-api v0.16.16-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.6

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.6

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.6

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.6

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.6

replace k8s.io/kubectl => k8s.io/kubectl v0.16.6

replace k8s.io/kubelet => k8s.io/kubelet v0.16.6

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.6

replace k8s.io/metrics => k8s.io/metrics v0.16.6

replace k8s.io/node-api => k8s.io/node-api v0.16.6

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.6

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.16.6

replace k8s.io/sample-controller => k8s.io/sample-controller v0.16.6

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20161114122254-48702e0da86b

replace github.com/containerd/cgroups => github.com/containerd/cgroups v0.0.0-20190717030353-c4b9ac5c7601
