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

package atzerolog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/model"
	"go.atatus.com/agent/module/atzerolog"
	"go.atatus.com/agent/transport/transporttest"
)

func ExampleWriter() {
	zerolog.ErrorStackMarshaler = atzerolog.MarshalErrorStack

	logger := zerolog.New(zerolog.MultiLevelWriter(os.Stdout, &atzerolog.Writer{}))
	logger.Error().Msg("boom")
}

func TestWriter(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	t0 := time.Unix(0, 0).UTC()
	zerolog.TimestampFunc = func() time.Time { return t0 }
	defer func() {
		zerolog.TimestampFunc = time.Now
	}()

	writer := &atzerolog.Writer{Tracer: tracer}
	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Error().Int("foo", 123).Msg("¡hola, mundo!")

	tracer.Flush(nil)
	payloads := transport.Payloads()
	assert.Len(t, payloads.Errors, 1)

	err0 := payloads.Errors[0]
	assert.Equal(t, "¡hola, mundo!", err0.Log.Message)
	assert.Equal(t, "error", err0.Log.Level)
	assert.Equal(t, "", err0.Log.LoggerName)
	assert.Equal(t, "", err0.Log.ParamMessage)
	assert.Equal(t, "TestWriter", err0.Culprit)
	assert.NotEmpty(t, err0.Log.Stacktrace)
	assert.Equal(t, model.Time(time.Unix(0, 0).UTC()), err0.Timestamp)
	assert.Zero(t, err0.ParentID)
	assert.Zero(t, err0.TraceID)
	assert.Zero(t, err0.TransactionID)
}

func TestWriterTraceContext(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	writer := &atzerolog.Writer{Tracer: tracer}
	logger := zerolog.New(writer)

	tx := tracer.StartTransaction("name", "type")
	ctx := atatus.ContextWithTransaction(context.Background(), tx)
	span, ctx := atatus.StartSpan(ctx, "name", "type")
	logger = logger.Hook(atzerolog.TraceContextHook(ctx))
	logger.Error().Msg("¡hola, mundo!")
	span.End()
	tx.End()

	tracer.Flush(nil)
	payloads := transport.Payloads()
	assert.Len(t, payloads.Transactions, 1)
	assert.Len(t, payloads.Spans, 1)
	assert.Len(t, payloads.Errors, 1)

	err0 := payloads.Errors[0]
	assert.Equal(t, payloads.Spans[0].ID, err0.ParentID)
	assert.Equal(t, payloads.Transactions[0].TraceID, err0.TraceID)
	assert.Equal(t, payloads.Transactions[0].ID, err0.TransactionID)
}

func TestWriterNonError(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	writer := &atzerolog.Writer{Tracer: tracer}
	logger := zerolog.New(writer)
	logger.Info().Msg("¡hola, mundo!")

	tracer.Flush(nil)
	payloads := transport.Payloads()
	assert.Empty(t, payloads.Errors)
}

func TestWriterMinLevel(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	writer := &atzerolog.Writer{
		Tracer:   tracer,
		MinLevel: zerolog.FatalLevel,
	}
	logger := zerolog.New(writer)
	logger.Error().Msg("oy vey!")

	tracer.Flush(nil)
	payloads := transport.Payloads()
	assert.Empty(t, payloads.Errors)
}

func TestWriterWithError(t *testing.T) {
	// Use our own ErrorStackMarshaler implementation,
	// which records a fully qualified function name.
	zerolog.ErrorStackMarshaler = atzerolog.MarshalErrorStack
	defer func() {
		zerolog.ErrorStackMarshaler = nil
	}()

	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	writer := &atzerolog.Writer{Tracer: tracer}
	logger := zerolog.New(writer)
	logger.Error().Stack().Err(makeError()).Msg("nope nope nope")

	tracer.Flush(nil)
	payloads := transport.Payloads()
	require.Len(t, payloads.Errors, 1)

	err0 := payloads.Errors[0]
	assert.Equal(t, "kablamo", err0.Exception.Message)
	assert.Equal(t, "nope nope nope", err0.Log.Message)
	assert.Equal(t, "makeError", err0.Culprit)
	require.NotEmpty(t, err0.Log.Stacktrace)
	require.NotEmpty(t, err0.Exception.Stacktrace)
	assert.NotEqual(t, err0.Log.Stacktrace, err0.Exception.Stacktrace)
	assert.Equal(t, "makeError", err0.Exception.Stacktrace[0].Function)
	assert.Equal(t, "(*Writer).WriteLevel", err0.Log.Stacktrace[0].Function)

	assert.Equal(t, "go.atatus.com/agent/module/atzerolog_test", err0.Exception.Stacktrace[0].Module)
	assert.NotEmpty(t, err0.Exception.Stacktrace[0].AbsolutePath)
}

func TestWriterWithErrorPkgErrorsStackMarshaler(t *testing.T) {
	// Marshal stack trace using rs/zerolog/pkgerrors, which
	// records only the unqualified function name.
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	defer func() {
		zerolog.ErrorStackMarshaler = nil
	}()

	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	writer := &atzerolog.Writer{Tracer: tracer}
	logger := zerolog.New(writer)
	logger.Error().Stack().Err(makeError()).Msg("nope nope nope")

	tracer.Flush(nil)
	payloads := transport.Payloads()
	require.Len(t, payloads.Errors, 1)

	err0 := payloads.Errors[0]
	assert.Equal(t, "kablamo", err0.Exception.Message)
	assert.Equal(t, "nope nope nope", err0.Log.Message)
	assert.Equal(t, "makeError", err0.Culprit)
	require.NotEmpty(t, err0.Log.Stacktrace)
	require.NotEmpty(t, err0.Exception.Stacktrace)
	assert.NotEqual(t, err0.Log.Stacktrace, err0.Exception.Stacktrace)
	assert.Equal(t, "makeError", err0.Exception.Stacktrace[0].Function)
	assert.Equal(t, "(*Writer).WriteLevel", err0.Log.Stacktrace[0].Function)

	// pkgerrors does not encode the package path, nor the absolute path.
	assert.Equal(t, "", err0.Exception.Stacktrace[0].Module)
	assert.Equal(t, "", err0.Exception.Stacktrace[0].AbsolutePath)
}

func makeError() error {
	return errors.New("kablamo")
}

func TestWriterTracerClosed(t *testing.T) {
	tracer, _ := transporttest.NewRecorderTracer()
	tracer.Close() // close it straight away, writer should return immediately

	writer := &atzerolog.Writer{Tracer: tracer}
	logger := zerolog.New(writer)
	logger.Error().Msg("boom")
}

func TestWriterFatal(t *testing.T) {
	if os.Getenv("_INSIDE_TEST") == "1" {
		tracer, _ := atatus.NewTracer("", "")
		logger := zerolog.New(&atzerolog.Writer{Tracer: tracer})
		logger.Fatal().Msg("fatality!")
	}

	var recorder transporttest.RecorderTransport
	mux := http.NewServeMux()
	mux.HandleFunc("/intake/v2/events", func(w http.ResponseWriter, req *http.Request) {
		if err := recorder.SendStream(req.Context(), req.Body); err != nil {
			panic(err)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cmd := exec.Command(os.Args[0], "-test.run=^TestWriterFatal$")
	cmd.Env = append(os.Environ(),
		"_INSIDE_TEST=1",
		"ATATUS_SERVER_URL="+server.URL,
		"ATATUS_LOG_FILE=stderr",
		"ATATUS_LOG_LEVEL=debug",
	)
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	defer func() {
		if t.Failed() {
			t.Logf("%s", output)
		}
	}()

	payloads := recorder.Payloads()
	require.Len(t, payloads.Errors, 1)
	assert.Equal(t, "fatality!", payloads.Errors[0].Log.Message)
	assert.Equal(t, "fatal", payloads.Errors[0].Log.Level)
}
