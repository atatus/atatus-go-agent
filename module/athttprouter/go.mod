module go.atatus.com/agent/module/athttprouter

require (
	github.com/julienschmidt/httprouter v1.2.0
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
	go.atatus.com/agent/module/athttp v1.1.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
