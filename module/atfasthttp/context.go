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

//go:build go1.12
// +build go1.12

package atfasthttp // import "go.atatus.com/agent/module/atfasthttp"

import (
	"context"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/internal/apmcontext"
	"go.atatus.com/agent/module/athttp"
)

const txKey = "atfasthttp_transaction"

func init() {
	origTransactionFromContext := apmcontext.TransactionFromContext
	apmcontext.TransactionFromContext = func(ctx context.Context) interface{} {
		if tx, ok := ctx.Value(txKey).(*txCloser); ok {
			return tx.tx
		}

		return origTransactionFromContext(ctx)
	}

	origBodyCapturerFromContext := apmcontext.BodyCapturerFromContext
	apmcontext.BodyCapturerFromContext = func(ctx context.Context) interface{} {
		if tx, ok := ctx.Value(txKey).(*txCloser); ok {
			return tx.bc
		}

		return origBodyCapturerFromContext(ctx)
	}
}

func setRequestContext(ctx *fasthttp.RequestCtx, tracer *atatus.Tracer, tx *atatus.Transaction) (*atatus.BodyCapturer, error) {
	req := new(http.Request)
	if err := fasthttpadaptor.ConvertRequest(ctx, req, true); err != nil {
		return nil, err
	}

	bc := tracer.CaptureHTTPRequestBody(req)
	tx.Context.SetHTTPRequest(req)
	tx.Context.SetHTTPRequestBody(bc)

	return bc, nil
}

func setResponseContext(ctx *fasthttp.RequestCtx, tx *atatus.Transaction, bc *atatus.BodyCapturer) {
	statusCode := ctx.Response.Header.StatusCode()

	tx.Result = athttp.StatusCodeResult(statusCode)
	if !tx.Sampled() {
		return
	}

	headers := make(http.Header)
	ctx.Response.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)

		headers.Set(sk, sv)
	})

	tx.Context.SetHTTPResponseHeaders(headers)
	tx.Context.SetHTTPStatusCode(statusCode)

	return
}

// StartTransactionWithBody returns a new Transaction with name,
// created with tracer, and taking trace context from ctx.
//
// If the transaction is not ignored, the request and the request body
// capturer will be returned with the transaction added to its context.
func StartTransactionWithBody(
	ctx *fasthttp.RequestCtx, tracer *atatus.Tracer, name string,
) (*atatus.Transaction, *atatus.BodyCapturer, error) {
	traceContext, ok := getRequestTraceparent(ctx, athttp.W3CTraceparentHeader)
	if !ok {
		traceContext, ok = getRequestTraceparent(ctx, athttp.AtatusTraceparentHeader)
	}

	if ok {
		tracestateHeader := string(ctx.Request.Header.Peek(athttp.TracestateHeader))
		traceContext.State, _ = athttp.ParseTracestateHeader(strings.Split(tracestateHeader, ",")...)
	}

	tx := tracer.StartTransactionOptions(name, "request", atatus.TransactionOptions{TraceContext: traceContext})

	bc, err := setRequestContext(ctx, tracer, tx)
	if err != nil {
		tx.End()

		return nil, nil, err
	}

	ctx.SetUserValue(txKey, newTxCloser(tx, bc))

	return tx, bc, nil
}
