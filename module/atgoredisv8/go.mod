module go.atatus.com/agent/module/atgoredisv8

go 1.14

require (
	github.com/go-redis/redis/v8 v8.0.0-beta.2
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.1.0
)

replace go.atatus.com/agent => ../..
