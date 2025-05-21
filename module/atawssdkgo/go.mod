module go.atatus.com/agent/module/atawssdkgo

go 1.15

require (
	github.com/aws/aws-sdk-go v1.38.14
	github.com/stretchr/testify v1.7.0
	go.atatus.com/agent v1.2.0
	go.atatus.com/agent/module/athttp v1.2.0
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp
