module go.atatus.com/agent/module/atchiv5

require (
	github.com/go-chi/chi/v5 v5.0.2
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
	go.atatus.com/agent/module/athttp v1.1.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.14
