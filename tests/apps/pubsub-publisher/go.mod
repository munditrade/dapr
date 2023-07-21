module app

go 1.17

require (
	github.com/dapr/dapr v1.10.9
	github.com/gorilla/mux v1.8.0
	google.golang.org/grpc v1.52.3
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20230124163310-31e0e69b6fc2 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace k8s.io/client => github.com/kubernetes-client/go v0.0.0-20190928040339-c757968c4c36
