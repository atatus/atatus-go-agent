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

package atot // import "go.atatus.com/agent/module/atot"

import (
	"io"
	"net/http"
	"net/textproto"
	"time"

	opentracing "github.com/opentracing/opentracing-go"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/module/athttp"
)

// New returns a new opentracing.Tracer backed by the supplied
// Atatus tracer.
//
// By default, the returned tracer will use atatus.DefaultTracer.
// This can be overridden by using a WithTracer option.
// The option WithSpanRefValidator allows one to override the
// set of spans that are recorded. By default only child-of
// spans are recorded.
func New(opts ...Option) opentracing.Tracer {
	t := &otTracer{
		tracer:         atatus.DefaultTracer,
		isValidSpanRef: isChildOfSpanRef,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// otTracer is an opentracing.Tracer backed by an atatus.Tracer.
type otTracer struct {
	tracer         *atatus.Tracer
	isValidSpanRef SpanRefValidator
}

// StartSpan starts a new OpenTracing span with the given name and zero or more options.
func (t *otTracer) StartSpan(name string, opts ...opentracing.StartSpanOption) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return t.StartSpanWithOptions(name, sso)
}

// StartSpanWithOptions starts a new OpenTracing span with the given name and options.
func (t *otTracer) StartSpanWithOptions(name string, opts opentracing.StartSpanOptions) opentracing.Span {
	// Because the Context method can be called at any time after
	// the span is finished, we cannot pool the objects.
	otSpan := &otSpan{
		tracer: t,
		tags:   opts.Tags,
		ctx: spanContext{
			tracer:    t,
			startTime: opts.StartTime,
		},
	}
	if opts.StartTime.IsZero() {
		otSpan.ctx.startTime = time.Now()
	}

	var parentTraceContext atatus.TraceContext
	if parentCtx, ok := t.parentSpanContext(opts.References); ok {
		if parentCtx.tx != nil && (parentCtx.tracer == t || parentCtx.tracer == nil) {
			opts := atatus.SpanOptions{
				Parent: parentCtx.traceContext, // parent span
				Start:  otSpan.ctx.startTime,
			}
			otSpan.span = parentCtx.tx.StartSpanOptions(name, "", opts)
			otSpan.ctx.tx = parentCtx.tx
			otSpan.ctx.traceContext = otSpan.span.TraceContext()
			return otSpan
		}
		parentTraceContext = parentCtx.traceContext
	}

	// There's no local parent context created by this tracer.
	otSpan.ctx.tx = t.tracer.StartTransactionOptions(name, "", atatus.TransactionOptions{
		TraceContext: parentTraceContext,
		Start:        otSpan.ctx.startTime,
	})
	otSpan.ctx.traceContext = otSpan.ctx.tx.TraceContext()
	return otSpan
}

func (t *otTracer) Inject(sc opentracing.SpanContext, format interface{}, carrier interface{}) error {
	spanContext, ok := sc.(*spanContext)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		writer, ok := carrier.(opentracing.TextMapWriter)
		if !ok {
			return opentracing.ErrInvalidCarrier
		}
		headerValue := athttp.FormatTraceparentHeader(spanContext.traceContext)
		writer.Set(athttp.W3CTraceparentHeader, headerValue)
		if t.tracer.ShouldPropagateLegacyHeader() {
			writer.Set(athttp.AtatusTraceparentHeader, headerValue)
		}
		if tracestate := spanContext.traceContext.State.String(); tracestate != "" {
			writer.Set(athttp.TracestateHeader, tracestate)
		}
		return nil
	case opentracing.Binary:
		writer, ok := carrier.(io.Writer)
		if !ok {
			return opentracing.ErrInvalidCarrier
		}
		return binaryInject(writer, spanContext.traceContext)
	default:
		return opentracing.ErrUnsupportedFormat
	}
}

func (t *otTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		var traceparentHeaderValue string
		var tracestateHeaderValues []string
		switch carrier := carrier.(type) {
		case opentracing.HTTPHeadersCarrier:
			traceparentHeaderValue = http.Header(carrier).Get(athttp.W3CTraceparentHeader)
			if traceparentHeaderValue == "" {
				traceparentHeaderValue = http.Header(carrier).Get(athttp.AtatusTraceparentHeader)
			}
			tracestateHeaderValues = http.Header(carrier)[athttp.TracestateHeader]
		case opentracing.TextMapReader:
			carrier.ForeachKey(func(key, val string) error {
				switch textproto.CanonicalMIMEHeaderKey(key) {
				case athttp.W3CTraceparentHeader:
					traceparentHeaderValue = val
				case athttp.AtatusTraceparentHeader:
					// The W3C header value always trumps the Elastic one,
					// hence we only set the value if not set already.
					if traceparentHeaderValue == "" {
						traceparentHeaderValue = val
					}
				case athttp.TracestateHeader:
					tracestateHeaderValues = append(tracestateHeaderValues, val)
				}
				return nil
			})
		default:
			return nil, opentracing.ErrInvalidCarrier
		}
		if traceparentHeaderValue == "" {
			return nil, opentracing.ErrSpanContextNotFound
		}
		traceContext, err := athttp.ParseTraceparentHeader(traceparentHeaderValue)
		if err != nil {
			return nil, err
		}
		traceContext.State, _ = athttp.ParseTracestateHeader(tracestateHeaderValues...)
		return &spanContext{tracer: t, traceContext: traceContext}, nil
	case opentracing.Binary:
		reader, ok := carrier.(io.Reader)
		if !ok {
			return nil, opentracing.ErrInvalidCarrier
		}
		traceContext, err := binaryExtract(reader)
		if err != nil {
			return nil, err
		}
		return &spanContext{tracer: t, traceContext: traceContext}, nil
	default:
		return nil, opentracing.ErrUnsupportedFormat
	}
}

// Option sets options for the OpenTracing Tracer implementation.
type Option func(*otTracer)

// WithTracer returns an Option which sets t as the underlying
// atatus.Tracer for constructing an OpenTracing Tracer.
func WithTracer(t *atatus.Tracer) Option {
	if t == nil {
		panic("t == nil")
	}
	return func(o *otTracer) {
		o.tracer = t
	}
}

// SpanRefValidator verifies if a span is valid and should be recorded.
type SpanRefValidator func(ref opentracing.SpanReference) bool

// WithSpanRefValidator returns an Option which sets the span validation
// function. By default only child-of span are considered valid.
func WithSpanRefValidator(validator SpanRefValidator) Option {
	if validator == nil {
		panic("validator == nil")
	}
	return func(o *otTracer) {
		o.isValidSpanRef = validator
	}
}

// TODO(axw) handle binary format once Trace-Context defines one.
// OpenTracing mandates that all implementations "MUST" support all
// of the builtin formats.

var (
	binaryInject  = binaryInjectUnsupported
	binaryExtract = binaryExtractUnsupported
)

func binaryInjectUnsupported(w io.Writer, traceContext atatus.TraceContext) error {
	return opentracing.ErrUnsupportedFormat
}

func binaryExtractUnsupported(r io.Reader) (atatus.TraceContext, error) {
	return atatus.TraceContext{}, opentracing.ErrUnsupportedFormat
}
