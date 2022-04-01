module go.atatus.com/agent/module/atrestful

require (
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/athttp v1.0.1
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
