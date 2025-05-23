module go.atatus.com/agent/module/atgoredis

go 1.12

require (
	github.com/go-redis/redis v6.15.3-0.20190424063336-97e6ed817821+incompatible
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.2.0
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace go.atatus.com/agent => ../..
