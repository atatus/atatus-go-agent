module go.atatus.com/agent/module/atgrpc

require (
	github.com/golang/protobuf v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/athttp v1.0.1
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	google.golang.org/grpc v1.17.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
