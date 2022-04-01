module go.atatus.com/agent/module/atot

require (
	github.com/opentracing/opentracing-go v1.1.0
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/athttp v1.0.1
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
