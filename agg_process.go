// Licensed to Atatus. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Atatus licenses this file to you under
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

package atatus

import (
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.atatus.com/agent/model"
	"go.atatus.com/agent/stacktrace"
)

func timeToMilliSeconds(t time.Time) int64 {
	return int64(t.UnixNano() / int64(time.Millisecond))
}

func timeDurationToMilliSeconds(d time.Duration) float64 {
	if d < 0 {
		return 0
	}
	return (float64(d/time.Microsecond) / 1000)
}

func roundThreeDecimals(val float64) (newVal float64) {
	return round(val, .5, 3)
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func buildAggTxn(tx *Transaction, td *TransactionData, analytics bool) (*aggTxn, *aggTxnAnalytics, int, *aggRequest) {
	var mt aggTxn
	var mta aggTxnAnalytics
	var agReq aggRequest
	mt.SetDuration(timeDurationToMilliSeconds(td.Duration))
	mt.Name = td.Name
	mt.Type = td.Type
	mt.Kind = golang

	mt.Layers = make(layerMap)

	statusCode := td.Context.response.StatusCode

	agReq = buildAggContextRequest(td.Context)

	// mt.BackgroundTxn

	if analytics == true {
		mta.RequestName = mt.aggTxnID.Name
		mta.Duration = mt.Durations[1]

		mta.CustomData = make(map[string]interface{})

		mta.TxnID = tx.traceContext.Span.String()
		mta.TraceID = tx.traceContext.Trace.String()

		mta.Timestamp = timeToMilliSeconds(td.timestamp)

		mta.UserID = td.Context.user.ID
		mta.UserName = td.Context.user.Username
		mta.UserEmail = td.Context.user.Email

		if td.Context.request.Body != nil {
			mta.RequestBody = td.Context.request.Body.Raw
		}

		mta.Method = td.Context.request.Method

		mta.StatusCode = statusCode

		if len(td.Context.request.Headers) > 0 {
			mta.RequestHeaders = make(map[string]interface{})
			for _, h := range td.Context.request.Headers {
				if len(h.Values) > 0 {
					mta.RequestHeaders[h.Key] = h.Values[0]

					if h.Key == "User-Agent" {
						mta.UserAgent = h.Values[0]
					}
				}
			}
		}

		{
			var u url.URL
			u.Host = td.Context.request.URL.Hostname
			if td.Context.request.URL.Port != "443" && td.Context.request.URL.Port != "80" {
				u.Host += ":" + td.Context.request.URL.Port
			}
			u.Scheme = td.Context.request.URL.Protocol

			u.Path = td.Context.request.URL.Path
			u.RawQuery = td.Context.request.URL.Search
			u.Fragment = td.Context.request.URL.Hash

			mta.URL = u.String()
		}

		if td.Context.request.Socket != nil {
			mta.IP = td.Context.request.Socket.RemoteAddress
		}

		if len(td.Context.response.Headers) > 0 {
			mta.ResponseHeaders = make(map[string]interface{})
			for _, h := range td.Context.response.Headers {
				if len(h.Values) > 0 {
					mta.ResponseHeaders[h.Key] = h.Values[0]
				}
			}
		}

		mta.Size = mta.TotalSize()
	}

	return &mt, &mta, statusCode, &agReq
}

func (agg *aggregator) processTxn(tx *Transaction, td *TransactionData) {

	if td == nil {
		return
	}

	analytics := agg.features.analytics
	if agg.config.APIAnalytics == false {
		analytics = false
	}

	mt, mta, statusCode, agReq := buildAggTxn(tx, td, analytics)

	txnSpans, ok := agg.b.txnSpan[tx.traceContext.Span.String()]
	if ok {
		for _, ts := range txnSpans {
			spanID := ts.layer.Key()
			span, ok := mt.Layers[spanID]
			if !ok {
				mt.Layers[spanID] = &ts.layer
			} else {
				span.Add(&ts.layer)
			}
		}
	}

	mt.SetGoCodeTiming()

	txnkey := mt.Key()
	txn, ok := agg.b.txn[txnkey]
	if !ok {
		agg.b.txn[txnkey] = mt
		txn = agg.b.txn[txnkey]
	} else {
		txn.Add(mt)
	}

	if mt.Durations[1] > float64(agg.config.TraceThreshold) {

		var trace aggTrace

		trace.aggTxnID = mt.aggTxnID
		trace.StartTime = timeToMilliSeconds(td.timestamp)
		trace.Duration = mt.Durations[1]
		trace.R = agReq
		trace.Entries = make([]aggTraceEntry, len(txnSpans))
		trace.Funcs = make([]string, 0)
		funcMap := make(map[string]int)
		for i, ts := range txnSpans {
			// spanID := ts.aggLayer.Key()
			trace.Entries[i].Level = 1

			if ts.Timestamp.After(td.timestamp) {
				trace.Entries[i].StartOffset = timeDurationToMilliSeconds(ts.Timestamp.Sub(td.timestamp))
			}
			trace.Entries[i].Duration = ts.Durations[1]
			trace.Entries[i].Layer.aggTxnID = ts.layer.aggTxnID

			index, ok := funcMap[ts.layer.aggTxnID.Name]
			if ok {
				trace.Entries[i].Index = index
			} else {
				trace.Funcs = append(trace.Funcs, ts.layer.aggTxnID.Name)
				index = len(trace.Funcs) - 1
				trace.Entries[i].Index = index
				funcMap[ts.layer.aggTxnID.Name] = index
			}
			if ts.Context.Database != nil {
				trace.Entries[i].Data = make(map[string]string)
				trace.Entries[i].Data["query"] = ts.Context.Database.Statement
				trace.Entries[i].Data["type"] = ts.Context.Database.Type
				trace.Entries[i].Data["user"] = ts.Context.Database.User
			}
		}

		agg.b.trace.Add(trace.aggTxnID.Key(), &trace)
	}

	// Custom
	// if msg.txnData.isBackgroundTxn() == false {
	if analytics == true {
		const txnAnalyticsAllowedCount int = 10000
		txnAnalyticsCount := len(agg.b.txnAnalytics)
		if txnAnalyticsCount < txnAnalyticsAllowedCount {
			agg.b.txnAnalytics = append(agg.b.txnAnalytics, mta)
		}
	}

	if statusCode >= 400 && statusCode != 404 {

		_, ok := agg.b.errMetric[txnkey]
		if !ok {
			var he httpErrorMetric
			he.aggTxnID = mt.aggTxnID
			he.StatusCode = make(map[string]int)
			agg.b.errMetric[txnkey] = he
		}

		errMetricString := strconv.Itoa(statusCode)
		agg.b.errMetric[txnkey].StatusCode[errMetricString]++

		httpErrorRequestCount := len(agg.b.errRequest)
		if httpErrorRequestCount <= 20 {
			var hr httpErrorRequest
			hr.aggTxnID = mt.aggTxnID
			hr.R = agReq
			// hr.Custom = msg.httpErrorData.Custom
			agg.b.errRequest = append(agg.b.errRequest, hr)
		}
	}

	td.reset(tx.tracer)

}

func mapSpanType(typ string) string {
	switch strings.ToLower(typ) {
	case "sql":
		return "SQL"
	case "mysql":
		return "MySQL"
	case "postgresql":
		return "Postgres"
	case "mssql":
		return "MS SQL"
	case "mongodb":
		return "MongoDB"
	case "redis":
		return "Redis"
	case "graphql":
		return "GraphQL"
	case "elasticsearch":
		return "Elasticsearch"
	case "cassandra":
		return "Cassandra"
	case "sqlite":
		return "SQLite"
	case "sqlite3":
		return "SQLite3"
	case "http":
		return "External Requests"
	case "https":
		return "External Requests"
	case "http2":
		return "External Requests"
	}

	return typ
}

func mapSpanKind(kind string) string {
	switch strings.ToLower(kind) {
	case "db":
		return "Database"
	case "cache":
		return "Database"
	case "ext":
		return "Remote"
	case "external":
		return "Remote"
	case "websocket":
		return "Remote"
	case "template":
		return "Template"
	}
	return kind
}

func buildAggSpan(s *Span, sd *SpanData) *aggLayer {
	var mp aggLayer

	mp.layer.SetDuration(timeDurationToMilliSeconds(sd.Duration))
	mp.layer.Name = sd.Name
	mp.layer.Type = mapSpanType(sd.Subtype)
	mp.layer.Kind = mapSpanKind(sd.Type)

	mp.Timestamp = sd.timestamp
	if sd.Context.model.Database != nil {
		mp.Context.Database = new(model.DatabaseSpanContext)
		mp.Context.Database.Instance = sd.Context.model.Database.Instance
		mp.Context.Database.Statement = sd.Context.model.Database.Statement
		mp.Context.Database.RowsAffected = new(int64)
		if sd.Context.model.Database.RowsAffected != nil {
			*mp.Context.Database.RowsAffected = *sd.Context.model.Database.RowsAffected
		} else {
			*mp.Context.Database.RowsAffected = 0
		}
		mp.Context.Database.Type = sd.Context.model.Database.Type
		mp.Context.Database.User = sd.Context.model.Database.User
	}

	return &mp
}

func (agg *aggregator) processSpan(s *Span, sd *SpanData) {

	if s.transactionID.Validate() == nil {
		mp := buildAggSpan(s, sd)
		txnid := s.transactionID.String()

		agg.b.txnSpan[txnid] = append(agg.b.txnSpan[txnid], mp)
	}

	sd.reset(s.tracer)
}

func buildAggContextRequest(ctx Context) (ar aggRequest) {

	ar.Method = ctx.request.Method

	if len(ctx.request.Headers) > 0 {
		for _, h := range ctx.request.Headers {
			if len(h.Values) > 0 {
				switch strings.ToLower(h.Key) {
				case "accept":
					ar.Accept = h.Values[0]
				case "accept-encoding":
					ar.AcceptEncoding = h.Values[0]
				case "accept-language":
					ar.AcceptLanguage = h.Values[0]
				case "referer":
					ar.Referer = h.Values[0]
				case "user-agent":
					ar.UserAgent = h.Values[0]
				}
			}
		}
	}

	ar.Host = ctx.request.URL.Hostname

	if len(ctx.request.URL.Port) > 0 {
		if i, err := strconv.Atoi(ctx.request.URL.Port); err == nil {
			ar.Port = i
		}
	}

	if ar.Port == 0 {
		if ctx.request.URL.Protocol == "http" {
			ar.Port = 80
		} else if ctx.request.URL.Protocol == "https" {
			ar.Port = 443
		}
	}

	ar.Path = ctx.request.URL.Path

	if ctx.request.Socket != nil {
		ar.IP = ctx.request.Socket.RemoteAddress
	}

	ar.StatusCode = ctx.response.StatusCode

	return ar
}

func buildAggRequest(req *http.Request) (ar aggRequest) {
	if req == nil {
		return
	}

	for k, v := range req.Header {
		if len(v) > 0 {
			switch strings.ToLower(k) {
			case "accept":
				ar.Accept = v[0]
			case "accept-encoding":
				ar.AcceptEncoding = v[0]
			case "accept-language":
				ar.AcceptLanguage = v[0]
			case "referer":
				ar.Referer = v[0]
			case "host":
				ar.Host = v[0]
			case "port":
				if i, err := strconv.Atoi(v[0]); err == nil {
					ar.Port = i
				}
			case "ip":
				ar.IP = v[0]
			case "method":
				ar.Method = v[0]
			case "useragent":
				ar.UserAgent = v[0]
			case "path":
				ar.Path = v[0]
			case "statuscode":
				if i, err := strconv.Atoi(v[0]); err == nil {
					ar.StatusCode = i
				}
			}
		}
	}

	return ar
}

func buildAggError(e *ErrorData) *aggError {
	var agErr aggError

	agErr.Timestamp = timeToMilliSeconds(e.Timestamp)

	agErr.TxnName = e.transactionName
	agErr.TxnType = e.transactionType
	agErr.TxnKind = golang
	// agErr.BackgroundTxn
	agErr.User.ID = e.Context.user.ID
	agErr.User.Name = e.Context.user.Username
	agErr.User.Email = e.Context.user.Email

	agErr.Request = buildAggContextRequest(e.Context)

	agErr.StackTraces = make([]stackTrace, 1)
	agErr.StackTraces[0].Message = e.exception.message

	agErr.StackTraces[0].Class = e.exception.Type.Name
	if agErr.StackTraces[0].Class == "" {
		agErr.StackTraces[0].Class = "Error"
	}
	agErr.StackTraces[0].Frames = make([]frame, 0)

	for _, v := range e.exception.stacktrace {
		var f frame

		var abspath string
		file := v.File
		if file != "" {
			if filepath.IsAbs(file) {
				abspath = file
			}
			file = filepath.Base(file)
		}

		f.Path = abspath
		f.File = file

		f.Method = v.Function
		f.LineNumber = v.Line

		packagePath, _ := stacktrace.SplitFunctionName(v.Function)
		f.InProject = stacktrace.IsLibraryPackage(packagePath)

		agErr.StackTraces[0].Frames = append(agErr.StackTraces[0].Frames, f)
	}
	// agErr.Custom

	return &agErr
}

func (agg *aggregator) processError(e *ErrorData) {

	agErr := buildAggError(e)
	if len(agg.b.err) <= 20 {
		agg.b.err = append(agg.b.err, agErr)
	}

	e.reset()
}

func buildAggMetrics(m *Metrics) map[string]model.Metric {
	am := make(map[string]model.Metric)
	for _, value := range m.metrics {
		for k, v := range value.Samples {
			am[k] = v
		}
	}
	return am
}

func (agg *aggregator) processMetrics(m *Metrics) {

	agMetric := buildAggMetrics(m)
	if len(agg.b.metrics) <= 20 {
		agg.b.metrics = append(agg.b.metrics, agMetric)
	}

	m.reset()
}
