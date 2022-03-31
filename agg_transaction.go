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
	"fmt"
	// "net"
	// "net/url"
	"strconv"
	"strings"
	"time"

	"go.atatus.com/agent/model"
)

type kind string

const (
	golang   string = "Go"
	database        = "Database"
	remote          = "Remote"
)

type aggTxnID struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Kind          string `json:"kind"`
	BackgroundTxn bool   `json:"background"`
}

type layer struct {
	aggTxnID
	Durations [4]float64 `json:"durations"`
}

type aggLayer struct {
	layer
	Timestamp time.Time
	Context   model.SpanContext
}

type layerMap map[string]*layer

type aggTxnLayerMap map[string][]*aggLayer

type aggTxn struct {
	layer
	Layers layerMap
}

type aggTxnAnalytics struct {
	IsAnalytics bool `json:"-"`
	Size        int  `json:"-"`

	Timestamp int64 `json:"timestamp"`

	TxnID   string `json:"txnId"`
	TraceID string `json:"traceId"`

	RequestName string `json:"name"`

	Duration float64 `json:"duration,omitempty"`

	StatusCode int    `json:"statusCode,omitempty"`
	Method     string `json:"method,omitempty"`
	URL        string `json:"url,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
	IP         string `json:"ip,omitempty"`

	RequestHeaders  map[string]interface{} `json:"requestHeaders,omitempty"`
	RequestBody     string                 `json:"requestBody,omitempty"`
	ResponseHeaders map[string]interface{} `json:"responseHeaders,omitempty"`
	ResponseBody    string                 `json:"responseBody,omitempty"`

	UserID            string `json:"userId,omitempty"`
	UserName          string `json:"userName,omitempty"`
	UserEmail         string `json:"userEmail,omitempty"`
	UserProperties    string `json:"userProperties,omitempty"`
	CompanyID         string `json:"companyId,omitempty"`
	CompanyProperties string `json:"companyProperties,omitempty"`

	CustomData map[string]interface{} `json:"customData,omitempty"`
}

func (r *aggTxnAnalytics) TotalSize() int {
	l := 8 + // len(r.Timestamp) +
		len(r.TxnID) +
		len(r.TraceID) +
		len(r.RequestName) +
		4 + // len(r.Duration) +
		4 + // len(r.DatabaseCount) +
		4 + // len(r.DatabaseDuration) +
		4 + // len(r.RemoteCount) +
		4 + // len(r.RemoteDuration) +
		2 + // len(r.StatusCode) +
		len(r.Method) +
		len(r.URL) +
		len(r.UserAgent) +
		len(r.IP) +
		len(r.RequestBody) +
		len(r.ResponseBody) +
		len(r.UserID) +
		len(r.CompanyID)

	for k, v := range r.RequestHeaders {
		l = l + len(k)
		vs, ok := v.(string)
		if ok {
			l = l + len(vs)
		} else {
			l = l + 4
		}
	}

	for k, v := range r.ResponseHeaders {
		l = l + len(k)
		vs, ok := v.(string)
		if ok {
			l = l + len(vs)
		} else {
			l = l + 4
		}
	}

	for k, v := range r.CustomData {
		l = l + len(k)
		vs, ok := v.(string)
		if ok {
			l = l + len(vs)
		} else {
			l = l + 4
		}
	}

	return l
}

type aggTxnMap map[string]*aggTxn

func (m *aggTxnID) Key() string {
	return m.Name + ":" + m.Type + ":" + string(m.Kind) + strconv.FormatBool(m.BackgroundTxn)
}

func (mp *layer) SetDuration(dur float64) {
	mp.Durations[0] = 1
	mp.Durations[1] = dur
	mp.Durations[2] = dur
	mp.Durations[3] = dur
}

func (mp *layer) SetAllValues(dur float64, min float64, max float64, count float64) {
	mp.Durations[0] = count
	mp.Durations[1] = dur
	mp.Durations[2] = min
	mp.Durations[3] = max
}

func (mp *layer) Add(amp *layer) {
	mp.Durations[0] += amp.Durations[0]
	mp.Durations[1] += amp.Durations[1]
	if mp.Durations[2] > amp.Durations[2] {
		mp.Durations[2] = amp.Durations[2]
	}
	if mp.Durations[3] < amp.Durations[3] {
		mp.Durations[3] = amp.Durations[3]
	}
}

func (mp *layer) String() string {
	return fmt.Sprintf("Name: %s, Type: %s, Kind: %s, Count: %g, Avg: %.7f, Min: %.7f, Max: %.7f\n", mp.Name, mp.Type, mp.Kind, mp.Durations[0], mp.Durations[1], mp.Durations[2], mp.Durations[3])
}

func (m *aggTxnID) isBackgroundTxn() bool {
	return m.BackgroundTxn
}

func (mp *aggTxn) isValidHistogramDuration() bool {
	return (mp.Durations[1] <= 150*1000.0)
}

func (mp *aggTxn) Add(amp *aggTxn) {
	mp.layer.Add(&amp.layer)
	for key := range amp.Layers {
		layer, ok := mp.Layers[key]
		if !ok {
			mp.Layers[key] = amp.Layers[key]
		} else {
			layer.Add(amp.Layers[key])
		}
	}
}

func handleNonSchemaURL(nonSchemaURL string) string {
	x := strings.Split(nonSchemaURL, "/")
	if len(x) > 0 {
		if x[0] == "" {
			return "(relative url)"
		}
		y := strings.Split(x[0], "@")
		if len(y) > 1 {
			return y[1]
		} else if len(y) > 0 {
			return y[0]
		}
	}
	return nonSchemaURL
}

func (mp *aggTxn) SetGoCodeTiming() {
	var layersDuration float64
	for _, value := range mp.Layers {
		layersDuration += value.Durations[1]
	}

	txnDuration := mp.Durations[1]

	if txnDuration < layersDuration {
		// errror
	} else if txnDuration > layersDuration {
		var languageLayer layer
		languageLayer.Name = mp.Type
		languageLayer.Type = golang
		languageLayer.Kind = golang
		languageLayer.SetDuration(roundThreeDecimals(txnDuration - layersDuration))
		key := languageLayer.Key()
		layer, ok := mp.Layers[key]
		if !ok {
			mp.Layers[key] = &languageLayer
		} else {
			layer.Add(&languageLayer)
		}
	}
}

type txnMetric struct {
	layer
	Layers []*layer `json:"traces"`
}

type txnPayload struct {
	header
	StartTime int64       `json:"startTime"`
	EndTime   int64       `json:"endTime"`
	T         []txnMetric `json:"transactions"`
}

type txnAnalyticsPayload struct {
	header
	StartTime int64              `json:"startTime"`
	EndTime   int64              `json:"endTime"`
	T         []*aggTxnAnalytics `json:"requests"`
}
