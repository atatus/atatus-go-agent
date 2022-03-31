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

//go:build go1.9
// +build go1.9

package atgorm // import "go.atatus.com/agent/module/atgorm"

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jinzhu/gorm"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/module/atsql"
)

const (
	apmContextKey = "atatus:context"
)

// WithContext returns a copy of db with ctx recorded for use by
// the callbacks registered via RegisterCallbacks.
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.Set(apmContextKey, ctx)
}

func scopeContext(scope *gorm.Scope) (context.Context, bool) {
	value, ok := scope.Get(apmContextKey)
	if !ok {
		return nil, false
	}
	ctx, _ := value.(context.Context)
	return ctx, ctx != nil
}

// RegisterCallbacks registers callbacks on db for reporting spans
// to Atatus. This is called automatically by atgorm.Open;
// it is provided for cases where a *gorm.DB is acquired by other
// means.
func RegisterCallbacks(db *gorm.DB) {
	registerCallbacks(db, atsql.DSNInfo{})
}

func registerCallbacks(db *gorm.DB, dsnInfo atsql.DSNInfo) {
	driverName := db.Dialect().GetName()
	switch driverName {
	case "postgres":
		driverName = "postgresql"
	}
	spanTypePrefix := fmt.Sprintf("db.%s.", driverName)
	querySpanType := spanTypePrefix + "query"
	execSpanType := spanTypePrefix + "exec"

	type params struct {
		spanType  string
		processor func() *gorm.CallbackProcessor
	}
	callbacks := map[string]params{
		"gorm:create": {
			spanType:  execSpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Create() },
		},
		"gorm:delete": {
			spanType:  execSpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Delete() },
		},
		"gorm:query": {
			spanType:  querySpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Query() },
		},
		"gorm:update": {
			spanType:  execSpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Update() },
		},
		"gorm:row_query": {
			spanType:  querySpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().RowQuery() },
		},
	}
	for name, params := range callbacks {
		const callbackPrefix = "atatus"
		params.processor().Before(name).Register(
			fmt.Sprintf("%s:before:%s", callbackPrefix, name),
			newBeforeCallback(params.spanType),
		)
		params.processor().After(name).Register(
			fmt.Sprintf("%s:after:%s", callbackPrefix, name),
			newAfterCallback(dsnInfo),
		)
	}
}

func newBeforeCallback(spanType string) func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		ctx, ok := scopeContext(scope)
		if !ok {
			return
		}
		span, ctx := atatus.StartSpan(ctx, "", spanType)
		if span.Dropped() {
			span.End()
			ctx = nil
		}
		scope.Set(apmContextKey, ctx)
	}
}

func newAfterCallback(dsnInfo atsql.DSNInfo) func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		ctx, ok := scopeContext(scope)
		if !ok {
			return
		}
		span := atatus.SpanFromContext(ctx)
		if span == nil {
			return
		}
		span.Name = atsql.QuerySignature(scope.SQL)
		span.Context.SetDestinationAddress(dsnInfo.Address, dsnInfo.Port)
		span.Context.SetDatabase(atatus.DatabaseSpanContext{
			Instance:  dsnInfo.Database,
			Statement: scope.SQL,
			Type:      "sql",
			User:      dsnInfo.User,
		})
		defer span.End()

		// Capture errors, except for "record not found", which may be expected.
		for _, err := range scope.DB().GetErrors() {
			if gorm.IsRecordNotFoundError(err) || err == sql.ErrNoRows {
				continue
			}
			if e := atatus.CaptureError(ctx, err); e != nil {
				e.Send()
			}
		}
	}
}