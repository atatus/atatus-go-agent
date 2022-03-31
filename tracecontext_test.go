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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	atatus "go.atatus.com/agent"
)

func TestTraceID(t *testing.T) {
	var id atatus.TraceID
	assert.EqualError(t, id.Validate(), "zero trace-id is invalid")

	id[0] = 1
	assert.NoError(t, id.Validate())
}

func TestSpanID(t *testing.T) {
	var id atatus.SpanID
	assert.EqualError(t, id.Validate(), "zero span-id is invalid")

	id[0] = 1
	assert.NoError(t, id.Validate())
}

func TestTraceOptions(t *testing.T) {
	opts := atatus.TraceOptions(0xFE)
	assert.False(t, opts.Recorded())

	opts = opts.WithRecorded(true)
	assert.True(t, opts.Recorded())
	assert.Equal(t, atatus.TraceOptions(0xFF), opts)

	opts = opts.WithRecorded(false)
	assert.False(t, opts.Recorded())
	assert.Equal(t, atatus.TraceOptions(0xFE), opts)
}

func TestTraceStateInvalidLength(t *testing.T) {
	const maxEntries = 32

	entries := make([]atatus.TraceStateEntry, 0, maxEntries)
	for i := 0; i < cap(entries); i++ {
		entries = append(entries, atatus.TraceStateEntry{Key: fmt.Sprintf("k%d", i), Value: "value"})
		ts := atatus.NewTraceState(entries...)
		assert.NoError(t, ts.Validate())
	}

	entries = append(entries, atatus.TraceStateEntry{Key: "straw", Value: "camel's back"})
	ts := atatus.NewTraceState(entries...)
	assert.EqualError(t, ts.Validate(), "tracestate contains more than the maximum allowed number of entries, 32")
}

func TestTraceStateDuplicateKey(t *testing.T) {
	// This test asserts:
	// 1. Accept a tracestate with duplicate keys.
	// 2. Use the last reference.
	// 3. Discard any duplicate 'at' entries.
	// 4. Keep any duplicate 3rd party system keys as is.
	ts := atatus.NewTraceState(
		atatus.TraceStateEntry{Key: "x", Value: "b"},
		atatus.TraceStateEntry{Key: "a", Value: "b"},
		atatus.TraceStateEntry{Key: "y", Value: "b"},
		atatus.TraceStateEntry{Key: "a", Value: "c"},
		atatus.TraceStateEntry{Key: "at", Value: "s:1;a:b"},
		atatus.TraceStateEntry{Key: "z", Value: "w"},
		atatus.TraceStateEntry{Key: "a", Value: "d"},
		atatus.TraceStateEntry{Key: "at", Value: "s:0.5;k:v"},
		atatus.TraceStateEntry{Key: "c", Value: "first"},
		atatus.TraceStateEntry{Key: "r", Value: "first"},
		atatus.TraceStateEntry{Key: "at", Value: "s:0.1;k:v"},
		atatus.TraceStateEntry{Key: "c", Value: "second"},
	)
	assert.NoError(t, ts.Validate())
	assert.Equal(t, "es=s:0.1;k:v,x=b,a=b,y=b,a=c,z=w,a=d,c=first,r=first,c=second", ts.String())
}

func TestTraceStateElasticEntryFirst(t *testing.T) {
	ts := atatus.NewTraceState(
		atatus.TraceStateEntry{Key: "at", Value: "s:1;a:b"},
		atatus.TraceStateEntry{Key: "z", Value: "w"},
		atatus.TraceStateEntry{Key: "a", Value: "d"},
	)
	assert.NoError(t, ts.Validate())
	assert.Equal(t, "es=s:1;a:b,z=w,a=d", ts.String())
}

func TestTraceStateInvalidKey(t *testing.T) {
	ts := atatus.NewTraceState(atatus.TraceStateEntry{Key: "~"})
	assert.EqualError(t, ts.Validate(), `invalid tracestate entry at position 0: invalid key "~"`)
}

func TestTraceStateInvalidValueLength(t *testing.T) {
	ts := atatus.NewTraceState(atatus.TraceStateEntry{Key: "oy"})
	assert.EqualError(t, ts.Validate(), `invalid tracestate entry at position 0: invalid value for key "oy": value is empty`)

	ts = atatus.NewTraceState(atatus.TraceStateEntry{Key: "oy", Value: strings.Repeat("*", 257)})
	assert.EqualError(t, ts.Validate(),
		`invalid tracestate entry at position 0: invalid value for key "oy": value contains 257 characters, maximum allowed is 256`)
}

func TestTraceStateInvalidValueCharacter(t *testing.T) {
	for _, value := range []string{
		string(rune(0)),
		"header" + string(rune(0)) + "trailer",
	} {
		ts := atatus.NewTraceState(atatus.TraceStateEntry{Key: "oy", Value: value})
		assert.EqualError(t, ts.Validate(),
			`invalid tracestate entry at position 0: invalid value for key "oy": value contains invalid character '\x00'`)
	}
}

func TestTraceStateInvalidElasticEntry(t *testing.T) {
	ts := atatus.NewTraceState(atatus.TraceStateEntry{Key: "at", Value: "foo"})
	assert.EqualError(t, ts.Validate(), `invalid tracestate entry at position 0: malformed 'at' tracestate entry`)

	ts = atatus.NewTraceState(atatus.TraceStateEntry{Key: "at", Value: "s:foo"})
	assert.EqualError(t, ts.Validate(), `invalid tracestate entry at position 0: strconv.ParseFloat: parsing "foo": invalid syntax`)

	ts = atatus.NewTraceState(atatus.TraceStateEntry{Key: "at", Value: "s:1.5"})
	assert.EqualError(t, ts.Validate(), `invalid tracestate entry at position 0: sample rate "1.5" out of range`)
}
