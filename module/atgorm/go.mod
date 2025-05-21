module go.atatus.com/agent/module/atgorm

require (
	cloud.google.com/go v0.40.0 // indirect
	github.com/jinzhu/gorm v1.9.10
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.2.0
	go.atatus.com/agent/module/atsql v1.2.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/atsql => ../atsql

go 1.13
