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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"go.atatus.com/agent/stacktrace"
)

const (
	hostinfoRelativePath     string = "/track/apm/hostinfo"
	errorRelativePath        string = "/track/apm/error"
	errorMetricRelativePath  string = "/track/apm/error_metric"
	txnRelativePath          string = "/track/apm/txn"
	traceRelativePath        string = "/track/apm/trace"
	metricsRelativePath      string = "/track/apm/metric"
	analyticsTxnRelativePath string = "/track/apm/analytics/txn"
)

func interfaceToJSONString(x interface{}) string {
	d, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return ""
	}
	return string(d)
}

type response400 struct {
	Blocked   bool   `json:"blocked"`
	Message   string `json:"errorMessage"`
	ErrorCode string `json:"errorCode"`
}

type hostinfoResponse200 struct {
	Analytics bool `json:"analytics"`

	CapturePercentiles    bool     `json:"capturePercentiles"`
	ExtRequestPatterns    []string `json:"extRequestPatterns"`
	IgnoreTxnNamePatterns []string `json:"ignoreTxnNamePatterns"`

	IgnoreHTTPFailuresPatterns map[string][]string `json:"ignoreHTTPFailurePatterns"`
	IgnoreExceptionPatterns    map[string][]string `json:"ignoreExceptionPatterns"`
}

type response struct {
	StatusCode          int
	Response400         response400
	HostinfoResponse200 hostinfoResponse200
}

var activeAggregator bool
var activeAggregatorCount int

func (agg *aggregator) sendToBackend(licenseKey, path string, d interface{}) (*response, error) {
	if agg.logger != nil {
		agg.logger.Debugf(interfaceToJSONString(d))
	}

	if path != hostinfoRelativePath {
		if activeAggregator == false {
			activeAggregator = true
			activeAggregatorCount = 0
		}
	}

	data, err := json.Marshal(d)
	if err != nil {
		if agg.logger != nil {
			agg.logger.Errorf("JSON Marshalling Failed: %s: %s\n", path, err.Error())
		}
		return nil, err
	}

	var host string
	if path == analyticsTxnRelativePath {
		if agg.config.NotifyHost == "https://apm-rx.atatus.com" ||
			agg.config.NotifyHost == "https://apm-rx-collector.atatus.com" {
			host = "https://an-rx.atatus.com"
		} else {
			host = agg.config.NotifyHost
		}
	} else {
		host = agg.config.NotifyHost
	}

	notifyHost := strings.TrimSuffix(host, "/") + path

	req, err := http.NewRequest("POST", notifyHost, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("licenseKey", agg.config.LicenseKey)
	q.Add("agent_name", agentLanguage)
	q.Add("agent_version", AgentVersion)
	req.URL.RawQuery = q.Encode()

	if agg.logger != nil {
		agg.logger.Debugf("Sending to URL: %+v\n", req.URL)
	}

	var client *http.Client

	if agg.config.NotifyProxy != "" {
		if agg.logger != nil {
			agg.logger.Debugf("Notify Proxy: %s\n", agg.config.NotifyProxy)
		}
		proxyURL, err := url.Parse(agg.config.NotifyProxy)
		if err != nil {
			if agg.logger != nil {
				agg.logger.Errorf("Notify Proxy Parsing Failed: %s Error: %s\n", agg.config.NotifyProxy, err.Error())
			}
			return nil, err
		}
		if agg.logger != nil {
			agg.logger.Debugf("Notify Proxy Details: %+v, Scheme: %s\n", proxyURL, proxyURL.Scheme)
		}

		var tr *http.Transport
		if proxyURL.Scheme == "http" || proxyURL.Scheme == "https" {
			tr = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
			client = &http.Client{Transport: tr}
		} else {
			client = &http.Client{}
		}

	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		if agg.logger != nil {
			agg.logger.Errorf("Sending request to %s failed with error: %s\n", path, err.Error())
		}
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var r response
	r.StatusCode = resp.StatusCode

	if path == hostinfoRelativePath && r.StatusCode == 200 {
		err := json.Unmarshal(body, &r.HostinfoResponse200)
		if err != nil {
			if agg.logger != nil {
				agg.logger.Errorf("Response status: %+v\n", resp.Status)
				agg.logger.Errorf("Response body: %+v\n", string(body))
				agg.logger.Errorf("Response JSON Unmarshaling Failed: %+v\n", err.Error())
			}
			return nil, err
		}
	} else if r.StatusCode == 400 {
		err := json.Unmarshal(body, &r.Response400)
		if err != nil {
			if agg.logger != nil {
				agg.logger.Errorf("Response status: %+v\n", resp.Status)
				agg.logger.Errorf("Response body: %+v\n", string(body))
				agg.logger.Errorf("Response JSON Unmarshaling Failed: %+v\n", err.Error())
			}
			return nil, err
		}
	}
	return &r, nil
}

var blocked bool

type agent struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type header struct {
	Agent          agent    `json:"agent"`
	Hostname       string   `json:"hostname"`
	UniqueHostname string   `json:"uniqueHostname"`
	ContainerID    string   `json:"containerId"`
	Tags           []string `json:"tags,omitempty"`
	AppVersion     string   `json:"version,omitempty"`
	ReleaseStage   string   `json:"releaseStage,omitempty"`
	AppName        string   `json:"appName,omitempty"`
	LicenseKey     string   `json:"licenseKey,omitempty"`
}

const (
	agentLanguage string = "Go"
)

func (agg *aggregator) flush(b *batchEvents) {

	lickey_appname_not_set := false

	if agg.config.LicenseKey == "" {
		lickey_appname_not_set = true
		if agg.logger != nil {
			agg.logger.Errorf("ATATUS_LICENSE_KEY is not set! Unable to send any data.")
		}
	}

	if agg.config.AppName == "" {
		lickey_appname_not_set = true
		if agg.logger != nil {
			agg.logger.Errorf("ATATUS_APP_NAME is not set! Unable to send any data.")
		}
	}

	if lickey_appname_not_set == true {
		if agg.logger != nil {
			agg.logger.Errorf("Set both ATATUS_LICENSE_KEY & ATATUS_APP_NAME for the APM to work!")
		}

		return
	}

	var h header
	h.Agent.Name = agentLanguage
	h.Agent.Version = AgentVersion
	h.Hostname = agg.config.Hostname
	h.AppName = agg.config.AppName
	h.AppVersion = agg.config.AppVersion
	h.ReleaseStage = agg.config.Environment
	h.LicenseKey = agg.config.LicenseKey
	if agg.host.hostID != "" {
		h.UniqueHostname = agg.host.hostID
	} else {
		h.UniqueHostname, _ = os.Hostname()
	}
	h.ContainerID = agg.host.dockerID
	if agg.config.Tags != nil {
		h.Tags = agg.config.Tags
	}
	// h.CustomData = agg.config.CustomData

	activeAggregatorCount++
	if activeAggregatorCount > (3600 / 30) {
		activeAggregator = false
	}

	if (agg.flushTickerCount%30) == 0 || agg.flushTickerCount == 1 { // represents 30 times * 60 seconds = 30 minutes
		var hp hostinfoPayload
		hp.header = h
		hp.hostinfo.Language = agentLanguage
		hp.hostinfo.Timestamp = time.Now().UnixNano() / 1000000
		hp.hostinfo.Environment.HostDetails = hostDetails
		hp.hostinfo.Environment.Setting = agg.agentSettingsMap()
		hp.hostinfo.Environment.GoLibrary = stacktrace.LibraryPackagesMap()
		hp.hostinfo.Active = activeAggregator

		r, err := agg.sendToBackend(agg.config.LicenseKey, hostinfoRelativePath, hp)
		agg.features.blocked = false
		agg.features.capturePercentiles = false
		agg.features.analytics = false
		if err == nil {
			if r.StatusCode == 400 {
				if agg.logger != nil {
					agg.logger.Errorf("Sending LicenseKey: %s Failed with StatusCode 400: %s\n", hp.header.LicenseKey, r.Response400.Message)
				}
				agg.features.blocked = r.Response400.Blocked
				if agg.features.blocked == true {
					if agg.logger != nil {
						agg.logger.Errorf("Atatus Blocked from Sending Data: %s\n", r.Response400.Message)
					}
				} else {
					if agg.logger != nil {
						agg.logger.Debugf("Setting Blocked to False: %+v\n", r)
					}
				}
			} else if r.StatusCode == 200 {
				agg.features.capturePercentiles = r.HostinfoResponse200.CapturePercentiles
				agg.features.analytics = r.HostinfoResponse200.Analytics
			}
		}
	}
	agg.flushTickerCount++

	if agg.features.blocked == true {
		return
	}

	if len(b.err) > 0 {
		var ep errPayload
		ep.header = h
		for _, v := range b.err {
			ep.E = append(ep.E, v)
		}
		agg.sendToBackend(agg.config.LicenseKey, errorRelativePath, ep)
	}

	if len(b.txn) > 0 {
		var tp txnPayload
		tp.EndTime = time.Now().UnixNano() / 1000000
		tp.StartTime = b.begin.UnixNano() / 1000000
		tp.header = h

		tp.T = make([]txnMetric, 0)
		for _, v := range b.txn {

			var t txnMetric

			t.layer = v.layer
			t.Layers = make([]*layer, len(v.Layers))
			i := 0
			for _, l := range v.Layers {
				t.Layers[i] = l
				i++
			}
			tp.T = append(tp.T, t)
		}
		if len(tp.T) > 0 {
			agg.sendToBackend(agg.config.LicenseKey, txnRelativePath, tp)
		}
	}

	if agg.features.analytics == true {
		if len(b.txnAnalytics) > 0 {
			var tp txnAnalyticsPayload
			tp.EndTime = time.Now().UnixNano() / 1000000
			tp.StartTime = b.begin.UnixNano() / 1000000
			tp.header = h

			totalSize := 0
			totalAnalyticsPayloadSize := 6 * 1024 * 1024
			startai := 0
			ai := 0
			if len(b.txnAnalytics) > 0 {
				for ai = range b.txnAnalytics {
					totalSize += b.txnAnalytics[ai].Size
					if totalSize >= totalAnalyticsPayloadSize {
						tp.T = b.txnAnalytics[startai:ai]
						if len(tp.T) > 0 {
							agg.sendToBackend(agg.config.LicenseKey, analyticsTxnRelativePath, tp)
						}
						startai = ai
						totalSize = b.txnAnalytics[ai].Size
					}
				}
				ai++
				if startai < ai {
					tp.T = b.txnAnalytics[startai:ai]
					if len(tp.T) > 0 {
						agg.sendToBackend(agg.config.LicenseKey, analyticsTxnRelativePath, tp)
					}
				}
			}
		}
	}

	if len(b.trace.PrimarySet.Traces) > 0 {
		numTraces := len(b.trace.PrimarySet.Traces)
		numBackgroundTraces := len(b.trace.BackgroundPrimarySet.Traces)
		if numTraces > 0 || numBackgroundTraces > 0 {
			var tp tracePayload
			tp.EndTime = time.Now().UnixNano() / 1000000
			tp.StartTime = b.begin.UnixNano() / 1000000
			tp.header = h

			// Copying Web Transactions Traces
			tp.T = b.trace.PrimarySet.Traces

			if numTraces < 5 {
				numSecondaryTraces := len(b.trace.SecondarySet.Traces)
				if numSecondaryTraces > 0 {
					sort.Sort(ByTraceDuration(b.trace.SecondarySet.Traces))
					fromSecondary := 5 - numTraces
					if fromSecondary >= numSecondaryTraces {
						fromSecondary = numSecondaryTraces
					}
					tp.T = append(tp.T, b.trace.SecondarySet.Traces[:fromSecondary]...)
				}
			}

			// Filling up Background Transactions after Web Transactions, if space is there
			numAvailableTraces := len(tp.T)
			if numAvailableTraces < 5 {
				numBackgroundPrimaryTraces := len(b.trace.BackgroundPrimarySet.Traces)
				if numBackgroundPrimaryTraces > 0 {
					sort.Sort(ByTraceDuration(b.trace.BackgroundPrimarySet.Traces))
					fromBackgroundPrimary := 5 - numAvailableTraces
					if fromBackgroundPrimary >= numBackgroundPrimaryTraces {
						fromBackgroundPrimary = numBackgroundPrimaryTraces
					}
					tp.T = append(tp.T, b.trace.BackgroundPrimarySet.Traces[:fromBackgroundPrimary]...)

					numAvailableTraces = len(tp.T)
					if numAvailableTraces < 5 {
						numBackgroundSecondaryTraces := len(b.trace.BackgroundSecondarySet.Traces)
						if numBackgroundSecondaryTraces > 0 {
							sort.Sort(ByTraceDuration(b.trace.BackgroundSecondarySet.Traces))
							fromBackgroundSecondary := 5 - numAvailableTraces
							if fromBackgroundSecondary >= numBackgroundSecondaryTraces {
								fromBackgroundSecondary = numBackgroundSecondaryTraces
							}
							tp.T = append(tp.T, b.trace.BackgroundSecondarySet.Traces[:fromBackgroundSecondary]...)
						}
					}
				}
			}

			for i := range tp.T {
				var singleTracePayload tracePayload
				singleTracePayload.EndTime = tp.EndTime
				singleTracePayload.StartTime = tp.StartTime
				singleTracePayload.header = tp.header
				singleTracePayload.T = append(singleTracePayload.T, tp.T[i])
				agg.sendToBackend(agg.config.LicenseKey, traceRelativePath, singleTracePayload)
			}
		}
	}

	if len(b.errMetric) > 0 {
		var emp errMetricPayload
		emp.EndTime = time.Now().UnixNano() / 1000000
		emp.StartTime = b.begin.UnixNano() / 1000000
		emp.header = h
		emp.M = make([]httpErrorMetric, 0)
		for _, v := range b.errMetric {
			emp.M = append(emp.M, v)
		}

		emp.R = make([]httpErrorRequest, 0)
		if len(b.errRequest) > 0 {
			emp.R = b.errRequest
		}

		agg.sendToBackend(agg.config.LicenseKey, errorMetricRelativePath, emp)
	}

	if len(b.metrics) > 0 {
		var mp metricsPayload
		mp.EndTime = time.Now().UnixNano() / 1000000
		mp.StartTime = b.begin.UnixNano() / 1000000
		mp.header = h
		mp.M = b.metrics

		agg.sendToBackend(agg.config.LicenseKey, metricsRelativePath, mp)
	}
}
