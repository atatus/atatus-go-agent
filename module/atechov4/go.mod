module go.atatus.com/agent/module/atechov4

require (
	github.com/labstack/echo/v4 v4.0.0
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
	go.atatus.com/agent/module/athttp v1.1.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
