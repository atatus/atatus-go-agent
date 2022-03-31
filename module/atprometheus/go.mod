module go.atatus.com/agent/module/atprometheus

require (
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.0
)

replace go.atatus.com/agent => ../..

go 1.13
