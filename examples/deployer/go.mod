module github.com/ease-lab/vhive/examples/deployer

go 1.16

replace github.com/ease-lab/vhive/examples/endpoint => ../endpoint

require (
	github.com/ease-lab/vhive/examples/endpoint v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.0
)
