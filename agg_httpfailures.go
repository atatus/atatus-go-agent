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
	"go.atatus.com/agent/model"
)

type httpErrorMetricMap map[string]httpErrorMetric

type httpErrorMetric struct {
	aggTxnID
	StatusCode map[string]int `json:"statusCodes"`
}

type httpErrorRequest struct {
	aggTxnID
	R      *aggRequest            `json:"request"`
	Custom map[string]interface{} `json:"customData,omitempty"`
}

type errMetricPayload struct {
	header
	StartTime int64              `json:"startTime"`
	EndTime   int64              `json:"endTime"`
	M         []httpErrorMetric  `json:"errorMetrics"`
	R         []httpErrorRequest `json:"errorRequests"`
}

type metricsPayload struct {
	header
	StartTime int64                     `json:"startTime"`
	EndTime   int64                     `json:"endTime"`
	M         []map[string]model.Metric `json:"golang"`
}
