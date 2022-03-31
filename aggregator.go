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

package atatus // import "go.atatus.com/agent"

import (
	"time"

	"go.atatus.com/agent/internal/apmlog"
	"go.atatus.com/agent/internal/sysinfo"
	"go.atatus.com/agent/model"
)

type batchEvents struct {
	txn          aggTxnMap
	txnSpan      aggTxnLayerMap
	txnAnalytics []*aggTxnAnalytics
	trace        *aggTraceBatch
	errMetric    httpErrorMetricMap
	errRequest   []httpErrorRequest
	err          []*aggError
	metrics      []map[string]model.Metric
	begin        time.Time
}

type aggChannels struct {
	txnChan     chan tracerEvent
	spanChan    chan tracerEvent
	errChan     chan tracerEvent
	metricsChan chan *Metrics
}

type hostInfo struct {
	hostID   string
	dockerID string
}

type features struct {
	blocked            bool
	capturePercentiles bool
	analytics          bool
}

type aggregator struct {
	config configuration

	b *batchEvents

	c aggChannels

	flushTicker      *time.Ticker
	flushTickerCount int

	process *model.Process
	system  *model.System
	host    hostInfo

	features features

	savedFramework string

	logger WarningLogger
}

func newBatch() *batchEvents {
	var b batchEvents
	b.begin = time.Now()

	b.txn = make(aggTxnMap)
	b.txnSpan = make(aggTxnLayerMap, 0)
	b.txnAnalytics = make([]*aggTxnAnalytics, 0)
	b.trace = new(aggTraceBatch)
	b.trace.Init()
	b.errMetric = make(httpErrorMetricMap)
	b.errRequest = make([]httpErrorRequest, 0)
	b.err = make([]*aggError, 0)
	b.metrics = make([]map[string]model.Metric, 0)
	return &b
}

var hostDetails map[string]interface{}

func newAggregator(service tracerService) *aggregator {

	config := newConfiguration(service)

	agg := aggregator{
		config:  config,
		process: &currentProcess,
		system:  &localSystem,
	}

	if apmlog.DefaultLogger != nil {
		agg.logger = apmlog.DefaultLogger
	}

	timeout := 10
	agg.flushTicker = time.NewTicker(time.Duration(timeout) * time.Second)

	agg.c.txnChan = make(chan tracerEvent, tracerEventChannelCap)
	agg.c.spanChan = make(chan tracerEvent, tracerEventChannelCap)
	agg.c.errChan = make(chan tracerEvent, tracerEventChannelCap)
	agg.c.metricsChan = make(chan *Metrics, tracerEventChannelCap)

	agg.savedFramework = ""

	hostDetails, _ = sysinfo.Host()

	if agg.host.hostID == "" {
		if val, ok := hostDetails["bootId"]; ok {
			if agg.logger != nil {
				agg.logger.Debugf("hostId is set from bootId")
			}
			agg.host.hostID = val.(string)
		}
	}
	if agg.host.hostID == "" {
		if val, ok := hostDetails["productId"]; ok {
			if agg.logger != nil {
				agg.logger.Debugf("hostId is set from productId")
			}
			agg.host.hostID = val.(string)
		}
	}
	if agg.host.hostID == "" {
		if val, ok := hostDetails["machineId"]; ok {
			if agg.logger != nil {
				agg.logger.Debugf("hostId is set from machineId")
			}
			agg.host.hostID = val.(string)
		}
	}
	if agg.host.hostID == "" {
		if val, ok := hostDetails["hostId"]; ok {
			if agg.logger != nil {
				agg.logger.Debugf("hostId is set from hostID")
			}
			agg.host.hostID = val.(string)
		}
	}
	if agg.host.hostID == "" {
		if agg.logger != nil {
			agg.logger.Debugf("hostId is set from hostname")
		}
		agg.host.hostID = config.Hostname
	}

	if agg.host.hostID == "" {
		if agg.logger != nil {
			agg.logger.Debugf("hostId is empty")
		}
	}

	if val, ok := hostDetails["containerId"]; ok {
		agg.host.dockerID = val.(string)
	}

	agg.b = newBatch()

	go agg.processEvents()

	return &agg
}

func (agg *aggregator) processEvents() {
	for {
		select {
		case event := <-agg.c.txnChan:
			agg.processTxn(event.tx.Transaction, event.tx.TransactionData)
		case event := <-agg.c.spanChan:
			agg.processSpan(event.span.Span, event.span.SpanData)
		case event := <-agg.c.errChan:
			agg.processError(event.err)
		case metrics := <-agg.c.metricsChan:
			agg.processMetrics(metrics)
		case <-agg.flushTicker.C:
			go agg.flush(agg.b)
			agg.b = newBatch()
		}
	}
}
