// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package atchi // import "go.atatus.com/agent/module/atchi"

import (
	"net/http"

	"github.com/go-chi/chi"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/module/athttp"
)

// Middleware returns a new chi middleware handler
// for tracing requests and reporting errors.
//
// The server request name will use the fully matched,
// parametrized route.
//
// By default, the middleware will use atatus.DefaultTracer.
// Use WithTracer to specify an alternative tracer.
func Middleware(o ...Option) func(http.Handler) http.Handler {
	opts := options{
		tracer:         atatus.DefaultTracer,
		requestIgnorer: athttp.DefaultServerRequestIgnorer(),
	}
	for _, o := range o {
		o(&opts)
	}
	return func(h http.Handler) http.Handler {
		return athttp.Wrap(
			h,
			athttp.WithTracer(opts.tracer),
			athttp.WithServerRequestName(routeRequestName),
			athttp.WithServerRequestIgnorer(opts.requestIgnorer),
		)
	}
}

func routeRequestName(r *http.Request) string {
	if routePattern, ok := getRoutePattern(r); ok {
		return r.Method + " " + routePattern
	}
	return athttp.UnknownRouteRequestName(r)
}

func getRoutePattern(r *http.Request) (string, bool) {
	routePath := r.URL.Path
	if r.URL.RawPath != "" {
		routePath = r.URL.RawPath
	}

	rctx := chi.RouteContext(r.Context())
	tctx := chi.NewRouteContext()
	if rctx.Routes.Match(tctx, r.Method, routePath) {
		return tctx.RoutePattern(), true
	}
	return "", false
}

type options struct {
	tracer         *atatus.Tracer
	requestIgnorer athttp.RequestIgnorerFunc
}

// Option sets options for tracing.
type Option func(*options)

// WithTracer returns an Option which sets t as the tracer
// to use for tracing server requests.
func WithTracer(t *atatus.Tracer) Option {
	if t == nil {
		panic("t == nil")
	}
	return func(o *options) {
		o.tracer = t
	}
}

// WithRequestIgnorer returns a Option which sets r as the
// function to use to determine whether or not a request should
// be ignored. If r is nil, all requests will be reported.
func WithRequestIgnorer(r athttp.RequestIgnorerFunc) Option {
	if r == nil {
		r = athttp.IgnoreNone
	}
	return func(o *options) {
		o.requestIgnorer = r
	}
}
