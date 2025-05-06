module go.atatus.com/agent/module/atfasthttp

go 1.13

require (
	github.com/stretchr/testify v1.6.1
	github.com/valyala/bytebufferpool v1.0.0
	github.com/valyala/fasthttp v1.26.0
	go.atatus.com/agent v1.1.0
	go.atatus.com/agent/module/athttp v1.1.0
)

replace (
	go.atatus.com/agent => ../..
	go.atatus.com/agent/module/athttp => ../athttp
)
