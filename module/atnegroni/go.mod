module go.atatus.com/agent/module/atnegroni

go 1.13

require (
	github.com/stretchr/testify v1.6.1
	github.com/urfave/negroni v1.0.0
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/athttp v1.0.1
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp
