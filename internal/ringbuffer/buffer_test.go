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

package ringbuffer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlockHeaderSize(t *testing.T) {
	size := binary.Size(BlockHeader{})
	assert.Equal(t, BlockHeaderSize, size)
}

func TestBuffer(t *testing.T) {
	b := New(150)
	assert.Equal(t, 0, b.Len())
	assert.Equal(t, 150, b.Cap())

	const block = `{"transaction":{"duration":0,"id":"00000000-0000-0000-0000-000000000000","name":"","timestamp":"0001-01-01T00:00:00Z","type":""}}`

	for i := 0; i < 10; i++ {
		b.WriteBlock([]byte(block), 0)
		blen := b.Len()
		assert.NotEqual(t, 0, blen)
		assert.Equal(t, 150, b.Cap())

		var bb bytes.Buffer
		_, n, err := b.WriteBlockTo(&bb)
		assert.NoError(t, err)
		assert.Equal(t, int64(blen-BlockHeaderSize), n)
		assert.Equal(t, block, bb.String())
		assert.Equal(t, 0, b.Len())
		_, n, err = b.WriteBlockTo(&bb)
		assert.Zero(t, n)
		assert.Equal(t, io.EOF, err)
	}
}

func TestBufferEviction(t *testing.T) {
	const block = `{"transaction":{"duration":0,"id":"00000000-0000-0000-0000-000000000000","name":"","timestamp":"0001-01-01T00:00:00Z","type":""}}`

	var evicted []BlockHeader
	b := New(300)
	b.Evicted = func(h BlockHeader) {
		evicted = append(evicted, h)
	}
	for i := 0; i < 100; i++ {
		b.WriteBlock([]byte(block), BlockTag(i))
	}
	assert.Equal(t, len(block)*2+2*BlockHeaderSize, b.Len())

	for i := 0; i < 2; i++ {
		var bb bytes.Buffer
		b.WriteBlockTo(&bb)
		assert.Equal(t, block, bb.String())
	}
	assert.Equal(t, 0, b.Len())
	assert.Len(t, evicted, 98)
	for i, h := range evicted {
		assert.Equal(t, BlockTag(i), h.Tag)
		assert.Equal(t, uint32(len(block)), h.Size)
	}
}

// TestConcurrentWriteBlockRace reproduces the customer-reported "agent stops
// sending data" bug: the tracer loop's WriteBlockTo and the aggregator's
// WriteBlock run on separate goroutines against the same *Buffer with no
// synchronization (see GO_BUGFIX_PLAN.md). A small buffer maximizes how often
// the read/write positions cross paths, so this reliably surfaces the race
// within a couple seconds instead of requiring hours of production traffic.
//
// Expected today (unfixed): this test panics or `go test -race` reports a
// DATA RACE. After the real fix (mutex + bounds check), it should pass clean
// under -race.
func TestConcurrentWriteBlockRace(t *testing.T) {
	b := New(4096)
	data := []byte(strings.Repeat("x", 200))

	var panicMsg atomic.Value // string
	deadline := time.Now().Add(3 * time.Second)

	run := func(who string, fn func()) {
		defer func() {
			if r := recover(); r != nil {
				panicMsg.CompareAndSwap(nil, fmt.Sprintf("%s panicked: %v", who, r))
			}
		}()
		for time.Now().Before(deadline) {
			fn()
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		run("WriteBlock", func() { b.WriteBlock(data, 1) })
	}()
	go func() {
		defer wg.Done()
		run("WriteBlockTo", func() { b.WriteBlockTo(ioutil.Discard) })
	}()
	wg.Wait()

	if v := panicMsg.Load(); v != nil {
		t.Fatalf("reproduced the ring buffer race from GO_BUGFIX_PLAN.md: %s", v)
	}
	t.Log("no panic in this run — the race is timing-dependent; re-run, or run with -race to catch the underlying data race directly")
}

func BenchmarkWrite(b *testing.B) {
	data := []byte(strings.Repeat("*", 1024))
	buf := New(10 * 1024 * 1024)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err := buf.WriteBlock(data[:], 0)
		if err != nil {
			panic(err)
		}
		b.SetBytes(int64(n))
	}
}

func BenchmarkWriteBlockTo(b *testing.B) {
	data := []byte(strings.Repeat("*", 300))
	buf := New(b.N * (len(data) + BlockHeaderSize))
	for i := 0; i < b.N; i++ {
		buf.WriteBlock(data[:], 0)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, n, err := buf.WriteBlockTo(ioutil.Discard)
		if err != nil {
			panic(err)
		}
		b.SetBytes(n)
	}
}
