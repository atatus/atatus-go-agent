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

package sysinfo

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

// Host provides all the host information that we need to identify the host
func Host() (map[string]interface{}, error) {

	var errMessage string
	h := make(map[string]interface{})

	m, err := mem.VirtualMemory()
	if err != nil {
		errMessage = errMessage + fmt.Sprintf("\nMemory: Unable to get memory details: %s", err.Error())
	} else {
		h["ram"] = m.Total
	}

	c, err := cpu.Info()
	if err != nil {
		errMessage = errMessage + fmt.Sprintf("\nCPU: Unable to get cpu details: %s", err.Error())
	} else {
		if len(c) > 0 {
			type cpuStat struct {
				Cores     int32   `json:"cores"`
				ModelName string  `json:"model"`
				Mhz       float64 `json:"mhz"`
			}
			cs := make([]cpuStat, len(c))
			for i := range c {
				cs[i].Cores = c[i].Cores
				cs[i].ModelName = c[i].ModelName
				cs[i].Mhz = c[i].Mhz
			}
			h["cpu"] = cs
		}
	}

	hi, err := host.Info()
	if err != nil {
		errMessage = errMessage + fmt.Sprintf("\nHostInfo: Unable to get host details: %s", err.Error())
	} else {
		if (hi.Hostname) != "" {
			h["hostname"] = hi.Hostname
		}
		if (hi.OS) != "" {
			h["os"] = hi.OS
		}
		if (hi.Platform) != "" && (hi.Platform) != (hi.OS) {
			h["platform"] = hi.Platform
		}
		if (hi.PlatformFamily) != "" && (hi.PlatformFamily) != (hi.Platform) && (hi.PlatformFamily) != (hi.OS) {
			h["family"] = hi.PlatformFamily
		}
		if hi.PlatformVersion != "" {
			h["version"] = hi.PlatformVersion
		}
		if (hi.KernelVersion) != "" {
			h["kernel"] = hi.KernelVersion
		}
		if (hi.VirtualizationSystem) != "" {
			h["vzsys"] = hi.VirtualizationSystem
		}
		if (hi.VirtualizationRole) != "" {
			h["vzrole"] = hi.VirtualizationRole
		}
		// We usually do not use hostId. We will have both BootID and UUID. If both BootID & UUID is not available, then we will get hostId
		if (hi.HostID) != "" {
			h["hostId"] = hi.HostID
		}
	}

	machineid, err := getMachineID()
	if err != nil {
		if err.Error() != "Not linux" {
			errMessage = errMessage + fmt.Sprintf("\nMachineID: Unable to get machineid: %s", err.Error())
		}
	} else {
		if machineid != "" {
			h["machineId"] = machineid
		}
	}

	uuid, err := getUUID()
	if err != nil {
		if err.Error() != "Not linux" {
			errMessage = errMessage + fmt.Sprintf("\nUUID: Unable to get uuid: %s", err.Error())
		}
	} else {
		if uuid != "" {
			h["productId"] = uuid
		}
	}

	bootid, err := getBootID()
	if err != nil {
		if err.Error() != "Not linux" {
			errMessage = errMessage + fmt.Sprintf("\nBootID: Unable to get bootid: %s", err.Error())
		}
	} else {
		if bootid != "" {
			h["bootId"] = bootid
		}
	}

	dockerID, err := getDockerID()
	if err != nil {
		if err.Error() != "Not linux" {
			// It need not be present always
		}
	} else {
		if dockerID != "" {
			h["containerId"] = dockerID
		}
	}
	// d, err := docker.GetDockerIDList()
	// if err != nil {
	// 	fmt.Printf("DockerIDList Failed: %s\n", err.Error())
	// } else {
	// 	fmt.Printf("Docker ID List\n")
	// 	for i := range d {
	// 		fmt.Printf("%s\n", d[i])
	// 	}
	// }

	// ds, err := docker.GetDockerStat()
	// if err != nil {
	// 	fmt.Printf("DockerStat Failed: %s\n", err.Error())
	// } else {
	// 	fmt.Printf("Docker Stat List\n")
	// 	for i := range ds {
	// 		fmt.Printf("%s\n", ds[i].String())
	// 	}
	// }

	if errMessage != "" {
		err = fmt.Errorf("%s", errMessage)
	}

	return h, err
}
