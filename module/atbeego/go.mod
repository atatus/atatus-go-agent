module go.atatus.com/agent/module/atbeego

require (
	github.com/astaxie/beego v1.11.1
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.2.0
	go.atatus.com/agent/module/athttp v1.2.0
	go.atatus.com/agent/module/atsql v1.2.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp

replace go.atatus.com/agent/module/atsql => ../atsql

go 1.13
