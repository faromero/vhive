module chained_function_eventing

go 1.16

replace github.com/ease-lab/vhive/utils/benchmarking/eventing => ../../../utils/benchmarking/eventing

require (
	github.com/ease-lab/vhive/utils/benchmarking/eventing v0.0.0-00010101000000-000000000000 // indirect
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/containerd/containerd v1.5.2
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)
