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

type plugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type hostEnv struct {
	Framework   string                 `json:"framework,omitempty"`
	GoLibrary   map[string]interface{} `json:"goLibraries"`
	Setting     map[string]interface{} `json:"settings"`
	HostDetails map[string]interface{} `json:"host"`
}

type hostinfo struct {
	Active      bool    `json:"active"`
	Timestamp   int64   `json:"timestamp"`
	Language    string  `json:"language"`
	Framework   string  `json:"framework,omitempty"`
	Environment hostEnv `json:"environment"`
}

type hostinfoPayload struct {
	header
	hostinfo
}

func (agg *aggregator) agentSettingsMap() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["appName"] = agg.config.AppName
	settings["appVersion"] = agg.config.AppVersion
	settings["agentVersion"] = AgentVersion
	settings["analytics"] = agg.config.APIAnalytics
	settings["environment"] = agg.config.Environment
	settings["goCompiler"] = goRuntime.Name
	settings["go"] = goRuntime.Version
	settings["traceThreshold"] = agg.config.TraceThreshold
	return settings
}
