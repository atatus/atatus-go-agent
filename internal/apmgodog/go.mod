module apmgodog

go 1.13

require (
	github.com/cucumber/godog v0.12.2
	go.atatus.com/agent v1.2.0
	go.atatus.com/agent/module/atgrpc v1.2.0
	go.atatus.com/agent/module/athttp v1.2.0
	go.elastic.co/fastjson v1.1.0
	google.golang.org/grpc v1.21.1
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/atgrpc => ../../module/atgrpc

replace go.atatus.com/agent/module/athttp => ../../module/athttp
