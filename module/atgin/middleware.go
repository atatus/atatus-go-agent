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

package atgin // import "go.atatus.com/agent/module/atgin"

import (
	"net/http"

	"github.com/gin-gonic/gin"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/module/athttp"
	"go.atatus.com/agent/stacktrace"
)

func init() {
	stacktrace.RegisterLibraryPackage(
		"github.com/gin-gonic",
		"github.com/gin-contrib",
	)
}

// Middleware returns a new Gin middleware handler for tracing
// requests and reporting errors.
//
// This middleware will recover and report panics, so it can
// be used instead of the standard gin.Recovery middleware.
//
// By default, the middleware will use atatus.DefaultTracer.
// Use WithTracer to specify an alternative tracer.
func Middleware(engine *gin.Engine, o ...Option) gin.HandlerFunc {
	m := &middleware{
		engine: engine,
		tracer: atatus.DefaultTracer,
	}
	for _, o := range o {
		o(m)
	}
	if m.requestIgnorer == nil {
		m.requestIgnorer = athttp.NewDynamicServerRequestIgnorer(m.tracer)
	}
	return m.handle
}

func (m *middleware) handle(c *gin.Context) {
	if !m.tracer.Recording() || m.requestIgnorer(c.Request) {
		c.Next()
		return
	}

	requestName := m.getRequestName(c)
	tx, body, req := athttp.StartTransactionWithBody(m.tracer, requestName, c.Request)
	defer tx.End()
	c.Request = req

	defer func() {
		if v := recover(); v != nil {
			if !c.Writer.Written() {
				c.AbortWithStatus(http.StatusInternalServerError)
			} else {
				c.Abort()
			}
			e := m.tracer.Recovered(v)
			e.SetTransaction(tx)
			setContext(&e.Context, c, body)
			e.Send()
		}
		c.Writer.WriteHeaderNow()
		tx.Result = athttp.StatusCodeResult(c.Writer.Status())

		if tx.Sampled() {
			setContext(&tx.Context, c, body)
		}

		for _, err := range c.Errors {
			e := m.tracer.NewError(err.Err)
			e.SetTransaction(tx)
			setContext(&e.Context, c, body)
			e.Handled = true
			e.Send()
		}
		body.Discard()
	}()
	c.Next()
}

func setContext(ctx *atatus.Context, c *gin.Context, body *atatus.BodyCapturer) {
	ctx.SetFramework("gin", gin.Version)
	ctx.SetHTTPRequest(c.Request)
	ctx.SetHTTPRequestBody(body)
	ctx.SetHTTPStatusCode(c.Writer.Status())
	ctx.SetHTTPResponseHeaders(c.Writer.Header())
}

// Option sets options for tracing.
type Option func(*middleware)

// WithTracer returns an Option which sets t as the tracer
// to use for tracing server requests.
func WithTracer(t *atatus.Tracer) Option {
	if t == nil {
		panic("t == nil")
	}
	return func(m *middleware) {
		m.tracer = t
	}
}

// WithRequestIgnorer returns a Option which sets r as the
// function to use to determine whether or not a request should
// be ignored. If r is nil, all requests will be reported.
func WithRequestIgnorer(r athttp.RequestIgnorerFunc) Option {
	if r == nil {
		r = athttp.IgnoreNone
	}
	return func(m *middleware) {
		m.requestIgnorer = r
	}
}
