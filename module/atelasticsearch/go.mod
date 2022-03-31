module go.atatus.com/agent/module/atelasticsearch

require (
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.0
	go.atatus.com/agent/module/athttp v1.0.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
