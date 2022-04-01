module go.atatus.com/agent/module/atecho

require (
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.2.8 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.6.1
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v0.0.0-20170224212429-dcecefd839c4 // indirect
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/athttp v1.0.1
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

go 1.13
