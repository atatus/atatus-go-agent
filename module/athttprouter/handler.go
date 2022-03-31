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

package athttprouter // import "go.atatus.com/agent/module/athttprouter"

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/module/athttp"
)

// Wrap wraps h such that it will report requests as transactions
// to Atatus, using route in the transaction name.
//
// By default, the returned Handle will use atatus.DefaultTracer.
// Use WithTracer to specify an alternative tracer.
//
// By default, the returned Handle will recover panics, reporting
// them to the configured tracer. To override this behaviour, use
// WithRecovery.
func Wrap(h httprouter.Handle, route string, o ...Option) httprouter.Handle {
	opts := gatherOptions(o...)
	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		if !opts.tracer.Recording() || opts.requestIgnorer(req) {
			h(w, req, p)
			return
		}
		tx, body, req := athttp.StartTransactionWithBody(opts.tracer, req.Method+" "+route, req)
		defer tx.End()

		w, resp := athttp.WrapResponseWriter(w)
		defer func() {
			if v := recover(); v != nil {
				if resp.StatusCode == 0 {
					w.WriteHeader(http.StatusInternalServerError)
				}
				opts.recovery(w, req, resp, body, tx, v)
			}
			athttp.SetTransactionContext(tx, req, resp, body)
			body.Discard()
		}()
		h(w, req, p)
		if resp.StatusCode == 0 {
			resp.StatusCode = http.StatusOK
		}
	}
}

// WrapNotFoundHandler wraps h so that it is traced. If h is nil, then http.NotFoundHandler() will be used.
func WrapNotFoundHandler(h http.Handler, o ...Option) http.Handler {
	if h == nil {
		h = http.NotFoundHandler()
	}
	return wrapHandlerUnknownRoute(h, o...)
}

// WrapMethodNotAllowedHandler wraps h so that it is traced. If h is nil, then a default handler
// will be used that returns status code 405.
func WrapMethodNotAllowedHandler(h http.Handler, o ...Option) http.Handler {
	if h == nil {
		h = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		})
	}
	return wrapHandlerUnknownRoute(h, o...)
}

func wrapHandlerUnknownRoute(h http.Handler, o ...Option) http.Handler {
	opts := gatherOptions(o...)
	return athttp.Wrap(
		h,
		athttp.WithTracer(opts.tracer),
		athttp.WithRecovery(opts.recovery),
		athttp.WithServerRequestName(athttp.UnknownRouteRequestName),
		athttp.WithServerRequestIgnorer(opts.requestIgnorer),
	)
}

func gatherOptions(o ...Option) options {
	opts := options{
		tracer: atatus.DefaultTracer,
	}
	for _, o := range o {
		o(&opts)
	}
	if opts.requestIgnorer == nil {
		opts.requestIgnorer = athttp.NewDynamicServerRequestIgnorer(opts.tracer)
	}
	if opts.recovery == nil {
		opts.recovery = athttp.NewTraceRecovery(opts.tracer)
	}
	return opts
}

type options struct {
	tracer         *atatus.Tracer
	recovery       athttp.RecoveryFunc
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

// WithRecovery returns an Option which sets r as the recovery
// function to use for tracing server requests.
func WithRecovery(r athttp.RecoveryFunc) Option {
	if r == nil {
		panic("r == nil")
	}
	return func(o *options) {
		o.recovery = r
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
