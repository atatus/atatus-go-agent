module genmod

require (
	github.com/pkg/errors v0.8.1
	go.atatus.com/agent v1.1.0
)

replace go.atatus.com/agent => ../..

go 1.13
