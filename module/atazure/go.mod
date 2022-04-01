module go.atatus.com/agent/module/atazure

go 1.14

require (
	github.com/Azure/azure-pipeline-go v0.2.3
	github.com/Azure/azure-storage-blob-go v0.14.0
	github.com/Azure/azure-storage-file-go v0.8.0
	github.com/Azure/azure-storage-queue-go v0.0.0-20191125232315-636801874cdd
	github.com/stretchr/testify v1.7.0
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/athttp v1.0.1
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
)

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../athttp
