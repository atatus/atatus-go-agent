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
	"html/template"
	"os"

	atatus "go.atatus.com/agent"
)

func ExampleTransaction_EnsureParent() {
	tx := atatus.DefaultTracer.StartTransactionOptions("name", "type", atatus.TransactionOptions{
		TraceContext: atatus.TraceContext{
			Trace: atatus.TraceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			Span:  atatus.SpanID{0, 1, 2, 3, 4, 5, 6, 7},
		},
	})
	defer tx.Discard()

	tpl := template.Must(template.New("").Parse(`
<script src="atatus-js-base/dist/bundles/atatus-js-base.umd.min.js"></script>
<script>
  atatus.init({
    serviceName: '',
    serverUrl: 'http://localhost:8200',
    pageLoadTraceId: {{.TraceContext.Trace}},
    pageLoadSpanId: {{.EnsureParent}},
    pageLoadSampled: {{.Sampled}},
  })
</script>
`))

	if err := tpl.Execute(os.Stdout, tx); err != nil {
		panic(err)
	}

	// Output:
	// <script src="atatus-js-base/dist/bundles/atatus-js-base.umd.min.js"></script>
	// <script>
	//   atatus.init({
	//     serviceName: '',
	//     serverUrl: 'http://localhost:8200',
	//     pageLoadTraceId: "000102030405060708090a0b0c0d0e0f",
	//     pageLoadSpanId: "0001020304050607",
	//     pageLoadSampled:  false ,
	//   })
	// </script>
}

func ExampleTransaction_EnsureParent_nilTransaction() {
	tpl := template.Must(template.New("").Parse(`
<script>
atatus.init({
  {{.TraceContext.Trace}},
  {{.EnsureParent}},
  {{.Sampled}},
})
</script>
`))

	// Demonstrate that Transaction's TraceContext, EnsureParent,
	// and Sampled methods will not panic when called with a nil
	// Transaction.
	tx := atatus.TransactionFromContext(context.Background())
	if err := tpl.Execute(os.Stdout, tx); err != nil {
		panic(err)
	}

	// Output:
	// <script>
	// atatus.init({
	//   "00000000000000000000000000000000",
	//   "0000000000000000",
	//    false ,
	// })
	// </script>
}
