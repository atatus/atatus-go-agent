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

package atatus_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/apmtest"
	"go.atatus.com/agent/model"
	"go.atatus.com/agent/transport"
	"go.atatus.com/agent/transport/transporttest"
)

func TestStartTransactionTraceContextOptions(t *testing.T) {
	testStartTransactionTraceContextOptions(t, false)
	testStartTransactionTraceContextOptions(t, true)
}

func testStartTransactionTraceContextOptions(t *testing.T, recorded bool) {
	tracer, _ := transporttest.NewRecorderTracer()
	defer tracer.Close()
	tracer.SetSampler(samplerFunc(func(atatus.TraceContext) bool {
		panic("nope")
	}))

	opts := atatus.TransactionOptions{
		TraceContext: atatus.TraceContext{
			Trace: atatus.TraceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			Span:  atatus.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
		},
	}
	opts.TraceContext.Options = opts.TraceContext.Options.WithRecorded(recorded)

	tx := tracer.StartTransactionOptions("name", "type", opts)
	result := tx.TraceContext()
	assert.Equal(t, recorded, result.Options.Recorded())
	tx.Discard()
}

func TestStartTransactionInvalidTraceContext(t *testing.T) {
	startTransactionInvalidTraceContext(t, atatus.TraceContext{
		// Trace is all zeroes, which is invalid.
		Span: atatus.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
	})
}

func startTransactionInvalidTraceContext(t *testing.T, traceContext atatus.TraceContext) {
	tracer, _ := transporttest.NewRecorderTracer()
	defer tracer.Close()

	var samplerCalled bool
	tracer.SetSampler(samplerFunc(func(atatus.TraceContext) bool {
		samplerCalled = true
		return true
	}))

	opts := atatus.TransactionOptions{TraceContext: traceContext}
	tx := tracer.StartTransactionOptions("name", "type", opts)
	assert.True(t, samplerCalled)
	tx.Discard()
}

func TestStartTransactionTraceParentSpanIDSpecified(t *testing.T) {
	startTransactionIDSpecified(t, atatus.TraceContext{
		Trace: atatus.TraceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		Span:  atatus.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
	})
}

func TestStartTransactionTraceIDSpecified(t *testing.T) {
	startTransactionIDSpecified(t, atatus.TraceContext{
		Trace: atatus.TraceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	})
}

func TestStartTransactionIDSpecified(t *testing.T) {
	startTransactionIDSpecified(t, atatus.TraceContext{})
}

func startTransactionIDSpecified(t *testing.T, traceContext atatus.TraceContext) {
	tracer, _ := transporttest.NewRecorderTracer()
	defer tracer.Close()

	opts := atatus.TransactionOptions{
		TraceContext:  traceContext,
		TransactionID: atatus.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
	}
	tx := tracer.StartTransactionOptions("name", "type", opts)
	assert.Equal(t, opts.TransactionID, tx.TraceContext().Span)
	tx.Discard()
}

func TestTransactionEnsureParent(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	tx := tracer.StartTransaction("name", "type")
	traceContext := tx.TraceContext()

	parentSpan := tx.EnsureParent()
	assert.NotZero(t, parentSpan)
	assert.NotEqual(t, traceContext.Span, parentSpan)

	// EnsureParent is idempotent.
	parentSpan2 := tx.EnsureParent()
	assert.Equal(t, parentSpan, parentSpan2)

	tx.End()

	// For an ended transaction, EnsureParent will return a zero value
	// even if the transaction had a parent at the time it was ended.
	parentSpan3 := tx.EnsureParent()
	assert.Zero(t, parentSpan3)

	tracer.Flush(nil)
	payloads := transport.Payloads()
	require.Len(t, payloads.Transactions, 1)
	assert.Equal(t, model.SpanID(parentSpan), payloads.Transactions[0].ParentID)
}

func TestTransactionParentID(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()

	tx := tracer.StartTransaction("name", "type")
	traceContext := tx.TraceContext()

	// root tx has no parent.
	parentSpan := tx.ParentID()
	assert.Zero(t, parentSpan)

	// Create a child transaction with the TraceContext from the parent.
	txChild := tracer.StartTransactionOptions("child", "type", atatus.TransactionOptions{
		TraceContext: tx.TraceContext(),
	})

	// Assert that the Parent ID isn't zero and matches the parent ID.
	// Parent TX ID
	parentTxID := traceContext.Span
	childSpanParentID := txChild.ParentID()
	assert.NotZero(t, childSpanParentID)
	assert.Equal(t, parentTxID, childSpanParentID)

	txChild.End()
	tx.End()

	// Assert that we can obtain the parent ID even after the transaction
	// has ended.
	assert.NotZero(t, txChild.ParentID())

	tracer.Flush(nil)
	payloads := tracer.Payloads()
	require.Len(t, payloads.Transactions, 2)

	// First recorded transaction Parent ID matches the child's Parent ID
	assert.Equal(t, model.SpanID(childSpanParentID), payloads.Transactions[0].ParentID)
	// First recorded transaction Parent ID matches the parent transaction ID.
	assert.Equal(t, model.SpanID(parentTxID), payloads.Transactions[0].ParentID)

	// Parent transaction has a zero ParentID.
	assert.Zero(t, payloads.Transactions[1].ParentID)
}

func TestTransactionParentIDWithEnsureParent(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()

	tx := tracer.StartTransaction("name", "type")

	rootParentIDEmpty := tx.ParentID()
	assert.Zero(t, rootParentIDEmpty)

	ensureParentResult := tx.EnsureParent()
	assert.NotZero(t, ensureParentResult)

	rootParentIDNotEmpty := tx.ParentID()
	assert.Equal(t, ensureParentResult, rootParentIDNotEmpty)

	tx.End()

	tracer.Flush(nil)
	payloads := tracer.Payloads()
	require.Len(t, payloads.Transactions, 1)

	assert.Equal(t, model.SpanID(rootParentIDNotEmpty), payloads.Transactions[0].ParentID)
}

func TestTransactionContextNotSampled(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()
	tracer.SetSampler(samplerFunc(func(atatus.TraceContext) bool { return false }))

	tx := tracer.StartTransaction("name", "type")
	tx.Context.SetLabel("foo", "bar")
	tx.End()
	tracer.Flush(nil)

	payloads := tracer.Payloads()
	require.Len(t, payloads.Transactions, 1)
	assert.Nil(t, payloads.Transactions[0].Context)
}

func TestTransactionNotRecording(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()
	tracer.SetRecording(false)
	tracer.SetSampler(samplerFunc(func(atatus.TraceContext) bool {
		panic("should not be called")
	}))

	tx := tracer.StartTransaction("name", "type")
	require.NotNil(t, tx)
	require.NotNil(t, tx.TransactionData)
	tx.End()
	require.Nil(t, tx.TransactionData)
	tracer.Flush(nil)

	payloads := tracer.Payloads()
	require.Empty(t, payloads.Transactions)
}

func TestTransactionSampleRate(t *testing.T) {
	type test struct {
		actualSampleRate   float64
		recordedSampleRate float64
		expectedTraceState string
	}
	tests := []test{
		{0, 0, "es=s:0"},
		{1, 1, "es=s:1"},
		{0.00001, 0.0001, "es=s:0.0001"},
		{0.55554, 0.5555, "es=s:0.5555"},
		{0.55555, 0.5556, "es=s:0.5556"},
		{0.55556, 0.5556, "es=s:0.5556"},
	}
	for _, test := range tests {
		test := test // copy for closure
		t.Run(fmt.Sprintf("%v", test.actualSampleRate), func(t *testing.T) {
			tracer := apmtest.NewRecordingTracer()
			defer tracer.Close()

			tracer.SetSampler(atatus.NewRatioSampler(test.actualSampleRate))
			tx := tracer.StartTransactionOptions("name", "type", atatus.TransactionOptions{
				// Use a known transaction ID for deterministic sampling.
				TransactionID: atatus.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
			})
			tx.End()
			tracer.Flush(nil)

			payloads := tracer.Payloads()
			assert.Equal(t, test.recordedSampleRate, *payloads.Transactions[0].SampleRate)
			assert.Equal(t, test.expectedTraceState, tx.TraceContext().State.String())
		})
	}
}

func TestTransactionUnsampledSampleRate(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()
	tracer.SetSampler(atatus.NewRatioSampler(0.5))

	// Create transactions until we get an unsampled one.
	//
	// Even though the configured sampling rate is 0.5,
	// we record sample_rate=0 to ensure the server does
	// not count the transaction toward metrics.
	var tx *atatus.Transaction
	for {
		tx = tracer.StartTransactionOptions("name", "type", atatus.TransactionOptions{})
		if !tx.Sampled() {
			tx.End()
			break
		}
		tx.Discard()
	}
	tracer.Flush(nil)

	payloads := tracer.Payloads()
	assert.Equal(t, float64(0), *payloads.Transactions[0].SampleRate)
	assert.Equal(t, "es=s:0", tx.TraceContext().State.String())
}

func TestTransactionSampleRatePropagation(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()

	for _, tracestate := range []atatus.TraceState{
		atatus.NewTraceState(atatus.TraceStateEntry{Keat "at", Value: "s:0.5"}),
		atatus.NewTraceState(atatus.TraceStateEntry{Key: "at", Value: "x:y;s:0.5;zz:y"}),
		atatus.NewTraceState(
			atatus.TraceStateEntry{Key: "other", Value: "s:1.0"},
			atatus.TraceStateEntry{Key: "at", Value: "s:0.5"},
		),
	} {
		tx := tracer.StartTransactionOptions("name", "type", atatus.TransactionOptions{
			TraceContext: atatus.TraceContext{
				Trace: atatus.TraceID{1},
				Span:  atatus.SpanID{1},
				State: tracestate,
			},
		})
		tx.End()
	}
	tracer.Flush(nil)

	payloads := tracer.Payloads()
	assert.Len(t, payloads.Transactions, 3)
	for _, tx := range payloads.Transactions {
		assert.Equal(t, 0.5, *tx.SampleRate)
	}
}

func TestTransactionSampleRateOmission(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()

	// For downstream transactions, sample_rate should be
	// omitted if a valid value is not found in tracestate.
	for _, tracestate := range []atatus.TraceState{
		atatus.TraceState{}, // empty
		atatus.NewTraceState(atatus.TraceStateEntry{Key: "other", Value: "s:1.0"}), // not "at", ignored
		atatus.NewTraceState(atatus.TraceStateEntry{Key: "at", Value: "s:123.0"}),  // out of range
		atatus.NewTraceState(atatus.TraceStateEntry{Key: "at", Value: ""}),         // 's' missing
		atatus.NewTraceState(atatus.TraceStateEntry{Key: "at", Value: "wat"}),      // malformed
	} {
		for _, sampled := range []bool{false, true} {
			tx := tracer.StartTransactionOptions("name", "type", atatus.TransactionOptions{
				TraceContext: atatus.TraceContext{
					Trace:   atatus.TraceID{1},
					Span:    atatus.SpanID{1},
					Options: atatus.TraceOptions(0).WithRecorded(sampled),
					State:   tracestate,
				},
			})
			tx.End()
		}
	}
	tracer.Flush(nil)

	payloads := tracer.Payloads()
	assert.Len(t, payloads.Transactions, 10)
	for _, tx := range payloads.Transactions {
		assert.Nil(t, tx.SampleRate)
	}
}

func TestTransactionDiscard(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	tx := tracer.StartTransaction("name", "type")
	tx.Discard()
	assert.Nil(t, tx.TransactionData)
	tx.End() // ending after discarding should be a no-op

	tracer.Flush(nil)
	payloads := transport.Payloads()
	require.Empty(t, payloads)
}

func TestTransactionDroppedSpansStats(t *testing.T) {
	exitSpanOpts := atatus.SpanOptions{ExitSpan: true}
	generateSpans := func(ctx context.Context, spans int) {
		for i := 0; i < spans; i++ {
			span, _ := atatus.StartSpanOptions(ctx,
				fmt.Sprintf("GET %d", i),
				fmt.Sprintf("request_%d", i),
				exitSpanOpts,
			)
			span.Duration = 10 * time.Microsecond
			span.End()
		}
	}
	type extraSpan struct {
		id, count int
	}
	generateExtraSpans := func(ctx context.Context, genExtra []extraSpan) {
		for _, extra := range genExtra {
			for i := 0; i < extra.count; i++ {
				span, _ := atatus.StartSpanOptions(ctx,
					fmt.Sprintf("GET %d", extra.id),
					fmt.Sprintf("request_%d", extra.id),
					exitSpanOpts,
				)
				span.Duration = 10 * time.Microsecond
				span.End()
			}
		}
	}
	// The default limit is 500 spans.
	// The default exit_span_min_duration is `1ms`.
	t.Run("DefaultLimit", func(t *testing.T) {
		tracer := apmtest.NewRecordingTracer()
		defer tracer.Close()

		tx, _, _ := tracer.WithTransaction(func(ctx context.Context) {
			generateSpans(ctx, 1000)
			generateExtraSpans(ctx, []extraSpan{
				{count: 100, id: 501},
				{count: 50, id: 600},
			})
		})
		// Ensure that the extra spans we generated are aggregated
		for _, span := range tx.DroppedSpansStats {
			if span.DestinationServiceResource == "request_501" {
				assert.Equal(t, 101, span.Duration.Count)
				assert.Equal(t, int64(1010), span.Duration.Sum.Us)
			} else if span.DestinationServiceResource == "request_600" {
				assert.Equal(t, 51, span.Duration.Count)
				assert.Equal(t, int64(510), span.Duration.Sum.Us)
			} else {
				assert.Equal(t, 1, span.Duration.Count)
				assert.Equal(t, int64(10), span.Duration.Sum.Us)
			}
		}
	})
	t.Run("DefaultLimit/DropShortExitSpans", func(t *testing.T) {
		tracer := apmtest.NewRecordingTracer()
		defer tracer.Close()
		// Set the exit span minimum duration. This test asserts that spans
		// with a duration over the span minimum duration are not dropped.
		tracer.SetExitSpanMinDuration(time.Microsecond)

		// Each of the generated spans duration is 10 microseconds.
		tx, spans, _ := tracer.WithTransaction(func(ctx context.Context) {
			generateSpans(ctx, 150)
		})

		require.Equal(t, 150, len(spans))
		require.Equal(t, 0, len(tx.DroppedSpansStats))
	})
	t.Run("MaxSpans100", func(t *testing.T) {
		tracer := apmtest.NewRecordingTracer()
		defer tracer.Close()
		// Assert that any spans over 100 are dropped and stats are aggregated.
		tracer.SetMaxSpans(100)

		tx, spans, _ := tracer.WithTransaction(func(ctx context.Context) {
			generateSpans(ctx, 300)
			generateExtraSpans(ctx, []extraSpan{
				{count: 50, id: 51},
				{count: 20, id: 60},
			})
		})

		require.Equal(t, 0, len(spans))
		require.Equal(t, 128, len(tx.DroppedSpansStats))

		for _, span := range tx.DroppedSpansStats {
			if span.DestinationServiceResource == "request_51" {
				assert.Equal(t, 51, span.Duration.Count)
				assert.Equal(t, int64(510), span.Duration.Sum.Us)
			} else if span.DestinationServiceResource == "request_60" {
				assert.Equal(t, 21, span.Duration.Count)
				assert.Equal(t, int64(210), span.Duration.Sum.Us)
			} else {
				assert.Equal(t, 1, span.Duration.Count)
				assert.Equal(t, int64(10), span.Duration.Sum.Us)
			}
		}
	})
	t.Run("MaxSpans10WithDisabledBreakdownMetrics", func(t *testing.T) {
		os.Setenv("ATATUS_BREAKDOWN_METRICS", "false")
		defer os.Unsetenv("ATATUS_BREAKDOWN_METRICS")
		tracer := apmtest.NewRecordingTracer()
		defer tracer.Close()

		// Assert that any spans over 10 are dropped and stats are aggregated.
		tracer.SetMaxSpans(10)

		// All spans except the one that we manually create will be dropped since
		// their duration is lower than `exit_span_min_duration`.
		tx, spans, _ := tracer.WithTransaction(func(ctx context.Context) {
			span, _ := atatus.StartSpanOptions(ctx, "name", "type", exitSpanOpts)
			span.Duration = time.Second
			span.End()
			generateSpans(ctx, 50)
		})

		require.Len(t, spans, 1)
		require.Len(t, tx.DroppedSpansStats, 50)

		// Ensure that the extra spans we generated are aggregated
		for _, span := range tx.DroppedSpansStats {
			assert.Equal(t, 1, span.Duration.Count)
			assert.Equal(t, int64(10), span.Duration.Sum.Us)
		}
	})
}

func TestTransactionOutcome(t *testing.T) {
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()

	tx1 := tracer.StartTransaction("name", "type")
	tx1.End()

	tx2 := tracer.StartTransaction("name", "type")
	tx2.Outcome = "unknown"
	tx2.End()

	tx3 := tracer.StartTransaction("name", "type")
	tx3.Context.SetHTTPStatusCode(400)
	tx3.End()

	tx4 := tracer.StartTransaction("name", "type")
	tx4.Context.SetHTTPStatusCode(500)
	tx4.End()

	tx5 := tracer.StartTransaction("name", "type")
	ctx := atatus.ContextWithTransaction(context.Background(), tx5)
	atatus.CaptureError(ctx, errors.New("an error")).Send()
	tx5.End()

	tracer.Flush(nil)
	transactions := tracer.Payloads().Transactions
	require.Len(t, transactions, 5)
	assert.Equal(t, "success", transactions[0].Outcome) // default
	assert.Equal(t, "unknown", transactions[1].Outcome) // specified
	assert.Equal(t, "success", transactions[2].Outcome) // HTTP status < 500
	assert.Equal(t, "failure", transactions[3].Outcome) // HTTP status >= 500
	assert.Equal(t, "failure", transactions[4].Outcome)
}

func BenchmarkTransaction(b *testing.B) {
	tracer, err := atatus.NewTracer("service", "")
	require.NoError(b, err)

	tracer.Transport = transport.Discard
	defer tracer.Close()

	names := []string{}
	for i := 0; i < 1000; i++ {
		names = append(names, fmt.Sprintf("/some/route/%d", i))
	}

	var mu sync.Mutex
	globalRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		mu.Lock()
		rand := rand.New(rand.NewSource(globalRand.Int63()))
		mu.Unlock()
		for pb.Next() {
			tx := tracer.StartTransaction(names[rand.Intn(len(names))], "type")
			tx.End()
		}
	})
}

type samplerFunc func(atatus.TraceContext) bool

func (f samplerFunc) Sample(t atatus.TraceContext) bool {
	return f(t)
}
