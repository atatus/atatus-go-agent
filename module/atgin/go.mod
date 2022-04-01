module go.atatus.com/agent/module/atgin

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/athttp v1.0.1
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
