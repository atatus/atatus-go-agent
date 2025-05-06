module go.atatus.com/agent/module/atgorilla

require (
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
	go.atatus.com/agent/module/athttp v1.1.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
