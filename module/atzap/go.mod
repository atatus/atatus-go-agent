module go.atatus.com/agent/module/atzap

require (
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.2.0
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1
)

replace go.atatus.com/agent => ../..

go 1.13
