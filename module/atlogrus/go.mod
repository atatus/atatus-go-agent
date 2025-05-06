module go.atatus.com/agent/module/atlogrus

require (
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.2.0
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
)

replace go.atatus.com/agent => ../..

go 1.13
