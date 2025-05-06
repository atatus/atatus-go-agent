module go.atatus.com/agent/module/atfiber

require (
	github.com/gofiber/fiber/v2 v2.18.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/valyala/fasthttp v1.29.0
	go.atatus.com/agent v1.1.0
	go.atatus.com/agent/module/atfasthttp v1.1.0
	go.atatus.com/agent/module/athttp v1.1.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

replace go.atatus.com/agent/module/atfasthttp => ../atfasthttp

go 1.13
