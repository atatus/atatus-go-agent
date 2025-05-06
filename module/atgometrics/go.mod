module go.atatus.com/agent/module/atgometrics

require (
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
)

replace go.atatus.com/agent => ../..

go 1.13
