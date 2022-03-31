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

//go:build linux
// +build linux

package sysinfo

import (
	"fmt"
	"regexp"
	"strings"
)

func getBootID() (string, error) {
	var lines []string
	var err error
	procBootID := hostProc("sys/kernel/random/boot_id")
	if pathExists(procBootID) == true {
		lines, err = readLines(procBootID)
		if err == nil && len(lines) > 0 && lines[0] != "" {
			return strings.ToLower(lines[0]), nil
		}
	} else {
		err = fmt.Errorf("Filename: \"%s\" does not exist", procBootID)
	}

	return "", err

}

func getUUID() (string, error) {
	var lines []string
	var err error
	sysProductUUID := hostSys("class/dmi/id/product_uuid")
	if pathExists(sysProductUUID) == true {
		lines, err = readLines(sysProductUUID)
		if err == nil && len(lines) > 0 && lines[0] != "" {
			return strings.ToLower(lines[0]), nil
		}
	} else {
		err = fmt.Errorf("Filename: \"%s\" does not exist", sysProductUUID)
	}

	return "", err
}

func getMachineID() (string, error) {
	var lines []string
	var err error
	machineID := hostEtc("machine-id")
	if pathExists(machineID) == true {
		lines, err = readLines(machineID)
		if err == nil && len(lines) > 0 && len(lines[0]) == 32 {
			st := lines[0]
			return fmt.Sprintf("%s-%s-%s-%s-%s", st[0:8], st[8:12], st[12:16], st[16:20], st[20:32]), nil
		}
	} else {
		err = fmt.Errorf("Filename: \"%s\" does not exist", machineID)
	}

	return "", err
}

// https://play.golang.org/p/iGXBnbg9Q2m
func getDockerID() (string, error) {
	cgroup_string := hostProc("self/cgroup")
	lines, err := readLines(cgroup_string)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(lines); i++ {
		match, err := regexp.MatchString("^\\d+:[^:]*?\\bcpu\\b[^:]*:", lines[i])
		if err != nil {
			return "", err
		}
		if match == true {
			re := regexp.MustCompile("([0-9a-f]{64})")
			split := strings.Split(lines[i], ":")
			if len(split) != 3 {
				return "", fmt.Errorf("Split length invalid: %d, %s", len(split), lines[i])
			}
			// fmt.Println("Matching Line-> " + lines[i])
			// fmt.Printf("%q\n", split)
			// fmt.Printf("%q\n", re.FindString(split[2]))
			return re.FindString(split[2]), nil
		}
	}

	return "", fmt.Errorf("Not found")
}
