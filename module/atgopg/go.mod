module go.atatus.com/agent/module/atgopg

require (
	github.com/go-pg/pg v8.0.4+incompatible
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/atsql v1.0.1
	mellium.im/sasl v0.2.1 // indirect
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/atsql => ../atsql

go 1.13
