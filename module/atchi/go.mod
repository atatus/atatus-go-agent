module go.atatus.com/agent/module/atchi

require (
	github.com/go-chi/chi v1.5.1
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.0
	go.atatus.com/agent/module/athttp v1.0.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
