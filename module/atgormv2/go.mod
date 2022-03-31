module go.atatus.com/agent/module/atgormv2

require (
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.0
	go.atatus.com/agent/module/atsql v1.0.0
	gorm.io/driver/mysql v1.0.2
	gorm.io/driver/postgres v1.0.2
	gorm.io/driver/sqlite v1.1.4-0.20200928065301-698e250a3b0d
	gorm.io/gorm v1.20.2
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/atsql => ../atsql

go 1.13
