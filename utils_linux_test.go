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

package atatus

import (
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCurrentProcessTitle(t *testing.T) {
	origProcessTitle, err := currentProcessTitle()
	assert.NoError(t, err)

	setProcessTitle := func(title string) {
		var buf [16]byte
		copy(buf[:], title)
		if _, _, errno := syscall.RawSyscall6(
			syscall.SYS_PRCTL, syscall.PR_SET_NAME,
			uintptr(unsafe.Pointer(&buf[0])),
			0, 0, 0, 0,
		); errno != 0 {
			t.Fatal(errno)
		}
	}

	const desiredTitle = "foo [bar]"
	setProcessTitle(desiredTitle)
	defer setProcessTitle(origProcessTitle)

	processTitle, err := currentProcessTitle()
	assert.NoError(t, err)
	assert.Equal(t, desiredTitle, processTitle)
}
