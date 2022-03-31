module go.atatus.com/agent/module/atrestfulv3

require (
	github.com/emicklei/go-restful/v3 v3.5.1
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.0
	go.atatus.com/agent/module/athttp v1.0.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
