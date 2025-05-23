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

package transport // import "go.atatus.com/agent/transport"

import (
	"context"
	"io"
)

// Transport provides an interface for sending streams of encoded model
// entities to the Atatus server, and for querying config. Methods
// are not required to be safe for concurrent use.
type Transport interface {
	// SendStream sends a data stream to the server, returning when the
	// stream has been closed (Read returns io.EOF) or the HTTP request
	// terminates.
	SendStream(context.Context, io.Reader) error
	SetNotifyURL(notifyHost, licenseKey, appName, agentVersion string) error // at_handling send stream
}
