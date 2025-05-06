module go.atatus.com/agent/module/atgokit

require (
	github.com/go-kit/kit v0.8.0
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
	go.atatus.com/agent/module/atgrpc v1.1.0
	go.atatus.com/agent/module/athttp v1.1.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	google.golang.org/grpc v1.17.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/atgrpc => ../atgrpc

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
