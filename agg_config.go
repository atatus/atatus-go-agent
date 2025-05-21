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
	"net/http"
	"os"
)

type configuration struct {
	// AppName holds the APM App name.
	//
	// If AppName is empty, the app name will be defined using the
	// ATATUS_APP_NAME environment variable, or if that is not set,
	// the executable name.
	AppName string

	// AppName holds the APM App version.
	//
	// If AppVersion is empty, the app version will be defined using the
	// ATATUS_APP_VERSION environment variable.
	AppVersion string

	// Environment holds the APM environment.
	//
	// If Environment is empty, the app version will be defined using the
	// ATATUS_ENVIRONMENT environment variable.
	Environment string

	// LicenseKey holds the APM License Key.
	//
	// If LicenseKey is empty, the license key will be defined using the
	// ATATUS_LICENSE_KEY environment variable.
	LicenseKey string

	// Analytics holds the APM Analytics Flag.
	//
	// If Analytics is empty, the API Analytics will be defined using the
	// ATATUS_ANALYTICS environment variable.
	Analytics bool

	// TraceThreshold holds the APM Trace Threshold value in ms.
	//
	// If TraceThreshold is empty, the TraceThreshold will be defined using the
	// ATATUS_TRACE_THRESHOLD environment variable.
	TraceThreshold int

	// Hostname represents the name of the server that you shall see in the Atatus UI. You can set it to a more
	// meaningful name, if you are running from containers
	//
	// Defaults to the hostname of the OS.
	Hostname string

	// Tags contains the default tags that shall be sent to all your errors and transactions
	//
	// Defaults to be empty
	Tags []string

	// CollectErrors controls if errors are to be collected and sent to Atatus servers
	//
	// Defaults to true, which enables errors to be collected
	CollectErrors bool

	// IgnoreStatusCodes is the list of HTTP status codes which are to be ignored from being
	// collected as errors. By default all HTTP status codes except 404 are captured as errors
	//
	// Default contains 404 in the ignore list
	IgnoreStatusCodes []int

	// CollectTransactions controls if transactions are to be collected and sent to Atatus servers
	//
	// Defaults to true, which enables transactions to be collected
	CollectTransactions bool

	// UseSSL controls whether the data is sent to Atatus server over http or https
	//
	// Defaults to true, which sends errors over https
	UseSSL bool

	// Transport can be configured on special environments like GAE or to configure proxy
	//
	// Defaults to the default http Transport
	Transport http.RoundTripper

	// NotifyHost controls the server path where the data is to be sent
	//
	// You can override using ATATUS_NOTIFY_HOST environment variable.
	NotifyHost string

	// NotifyProxy controls the proxy that has to be used
	//
	// Defaults to None
	NotifyProxy string

	// NotifyInterval controls the default interval on which the data is to be flushed
	//
	// Defaults to 60 seconds
	NotifyInterval int

	// Logger controls where the Atatus agent logs will be printed.
	//
	// Defaults to go's default logger
	// Logger Logger

	// SendCode controls if Atatus agent shall try to read local files and send the code in case of errors
	//
	// Defaults to true
	SendCode bool

	// Enabled controls if Atatus agent shall collect and send performance metrics
	//
	// Defaults to true
	Enabled bool
}

func newConfiguration(service tracerService) configuration {
	var c configuration

	c.AppName = service.AppName
	c.AppVersion = service.AppVersion
	c.Environment = service.Environment
	c.LicenseKey = service.LicenseKey
	c.Analytics = service.Analytics
	c.TraceThreshold = service.TraceThreshold
	c.NotifyHost = service.NotifyHost
	c.NotifyInterval = 60
	c.Hostname, _ = os.Hostname()
	c.CollectErrors = true
	c.IgnoreStatusCodes = []int{404}
	c.CollectTransactions = true
	c.UseSSL = true
	c.SendCode = true
	c.Enabled = true

	return c
}
