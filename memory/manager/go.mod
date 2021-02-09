module github.com/ease-lab/vhive/memory/manager

go 1.15

// Workaround for github.com/containerd/containerd issue #3031
replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible

replace github.com/ease-lab/vhive/metrics => ../../misc

require (
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/containerd/containerd v1.3.6
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b // indirect
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/containerd/typeurl v1.0.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/ease-lab/vhive/metrics v0.0.0-20201130161247-acbfdab4ba15
	github.com/ftrvxmtrx/fd v0.0.0-20150925145434-c6d800382fff
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.0.2 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd
	gonum.org/v1/gonum v0.7.0
)
