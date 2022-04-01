module go.atatus.com/agent/module/atelasticsearch/internal/integration

require (
	github.com/elastic/go-elasticsearch/v7 v7.5.0
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/mailru/easyjson v0.0.0-20180823135443-60711f1a8329 // indirect
	github.com/olivere/elastic v6.2.16+incompatible
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.0.1
	go.atatus.com/agent/module/atelasticsearch v1.0.1
)

replace go.atatus.com/agent => ../../../..

replace go.atatus.com/agent/module/atelasticsearch => ../..

replace go.atatus.com/agent/module/athttp => ../../../athttp

go 1.13
