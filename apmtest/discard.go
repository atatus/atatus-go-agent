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

package apmtest // import "go.atatus.com/agent/apmtest"

import (
	"log"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/transport/transporttest"
)

// DiscardTracer is an atatus.Tracer that discards all events.
//
// This tracer may be used by multiple tests, and so should
// not be modified or closed.
//
// Importing apmttest will close atatus.DefaultTracer, and update
// it to this value.
var DiscardTracer *atatus.Tracer

// NewDiscardTracer returns a new atatus.Tracer that discards all events.
func NewDiscardTracer() *atatus.Tracer {
	tracer, err := atatus.NewTracerOptions(atatus.TracerOptions{
		Transport: transporttest.Discard,
	})
	if err != nil {
		log.Fatal(err)
	}
	return tracer
}

func init() {
	atatus.DefaultTracer.Close()
	tracer := NewDiscardTracer()
	DiscardTracer = tracer
	atatus.DefaultTracer = DiscardTracer
}
