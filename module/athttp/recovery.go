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

package athttp // import "go.atatus.com/agent/module/athttp"

import (
	"net/http"

	atatus "go.atatus.com/agent"
)

// RecoveryFunc is the type of a function for use in WithRecovery.
type RecoveryFunc func(
	w http.ResponseWriter,
	req *http.Request,
	resp *Response,
	body *atatus.BodyCapturer,
	tx *atatus.Transaction,
	recovered interface{},
)

// NewTraceRecovery returns a RecoveryFunc for use in WithRecovery.
//
// The returned RecoveryFunc will report recovered error to Atatus
// using the given Tracer, or atatus.DefaultTracer if t is nil. The
// error will be linked to the given transaction.
//
// If headers have not already been written, a 500 response will be sent.
func NewTraceRecovery(t *atatus.Tracer) RecoveryFunc {
	if t == nil {
		t = atatus.DefaultTracer
	}
	return func(
		w http.ResponseWriter,
		req *http.Request,
		resp *Response,
		body *atatus.BodyCapturer,
		tx *atatus.Transaction,
		recovered interface{},
	) {
		e := t.Recovered(recovered)
		e.SetTransaction(tx)
		SetContext(&e.Context, req, resp, body)
		e.Send()
	}
}
