module tracecontexttest

require go.atatus.com/agent/module/athttp v1.0.0

replace go.atatus.com/agent => ../..

replace go.atatus.com/agent/module/athttp => ../../module/athttp

go 1.13
