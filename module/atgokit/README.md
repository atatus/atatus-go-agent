# atgokit

Package atgokit provides examples and integration tests
for tracing services implemented with Go kit.

We do not provide any Go kit specific code, as the other
generic modules ([module/athttp](../athttp) and [module/atgrpc](../atgrpc))
are sufficient.

Go kit-based HTTP servers can be traced by instrumenting the
kit/transport/http.Server with athttp.Wrap, and HTTP clients
can be traced by providing a net/http.Client instrumented with
athttp.WrapClient.

Go kit-based gRPC servers and clients can both be wrapped using
the interceptors provided in module/atgrpc.
