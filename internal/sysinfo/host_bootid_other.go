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

//go:build !linux
// +build !linux

package sysinfo

import (
	"fmt"
)

func getBootID() (string, error) {
	return "", fmt.Errorf("Not linux")
}

func getUUID() (string, error) {
	return "", fmt.Errorf("Not linux")
}

func getMachineID() (string, error) {
	return "", fmt.Errorf("Not linux")
}

func getDockerID() (string, error) {
	return "", fmt.Errorf("Not linux")
}
