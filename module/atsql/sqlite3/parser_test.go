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

package atsqlite3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "go.atatus.com/agent/apmtest" // disable default tracer
	"go.atatus.com/agent/module/atsql"
	atsqlite3 "go.atatus.com/agent/module/atsql/sqlite3"
)

func TestParseDSN(t *testing.T) {
	assert.Equal(t, atsql.DSNInfo{Database: "test.db"}, atsqlite3.ParseDSN("test.db"))
	assert.Equal(t, atsql.DSNInfo{Database: ":memory:"}, atsqlite3.ParseDSN(":memory:"))
	assert.Equal(t, atsql.DSNInfo{Database: "file:test.db"}, atsqlite3.ParseDSN("file:test.db?cache=shared&mode=memory"))
}
