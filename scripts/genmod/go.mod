module genmod

require (
	github.com/pkg/errors v0.8.1
	go.atatus.com/agent v1.0.1
)

replace go.atatus.com/agent => ../..

go 1.13
