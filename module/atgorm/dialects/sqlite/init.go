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

// Package atgormsqlite imports the gorm sqlite dialect package,
// and also registers the sqlite3 driver with atsql.
package atgormsqlite // import "go.atatus.com/agent/module/atgorm/dialects/sqlite"

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite" // import the sqlite dialect

	_ "go.atatus.com/agent/module/atsql/sqlite3" // register sqlite3 with atsql
)
