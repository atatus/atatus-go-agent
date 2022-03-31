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

import ()

type frame struct {
	File       string `json:"f,omitempty"`
	Path       string `json:"p,omitempty"`
	Method     string `json:"m,omitempty"`
	LineNumber int    `json:"ln,omitempty"`
	InProject  bool   `json:"inp,omitempty"`
}

type stackTrace struct {
	Class   string  `json:"class,omitempty"`
	Message string  `json:"message,omitempty"`
	Frames  []frame `json:"stacktrace"`
}

type aggRequest struct {
	Accept         string `json:"accept,omitempty"`
	AcceptEncoding string `json:"accept-encoding,omitempty"`
	AcceptLanguage string `json:"accept-language,omitempty"`
	Referer        string `json:"referer,omitempty"`
	Host           string `json:"host,omitempty"`
	Port           int    `json:"port,omitempty"`
	IP             string `json:"ip,omitempty"`
	Method         string `json:"method,omitempty"`
	UserAgent      string `json:"userAgent,omitempty"`
	Path           string `json:"path,omitempty"`
	StatusCode     int    `json:"statusCode,omitempty"`
}

type rxUser struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"userName,omitempty"`
	Email string `json:"userEmail,omitempty"`
}

type aggError struct {
	Timestamp     int64  `json:"timestamp"`
	TxnName       string `json:"transaction"`
	TxnType       string `json:"type"`
	TxnKind       string `json:"kind"`
	BackgroundTxn bool   `json:"background"`
	// Tags      []string `json:"tags"`
	User rxUser `json:"user,omitempty"`
	// CustomData  CustomData  `json:"customData"`
	// GroupingKey string `json:"groupingKey"`
	Request     aggRequest             `json:"request"`
	StackTraces []stackTrace           `json:"exceptions"`
	Custom      map[string]interface{} `json:"customData,omitempty"`
}

type errPayload struct {
	header
	E []*aggError `json:"errors"`
}
