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

type aggTraceLayer struct {
	aggTxnID
}

type aggTraceEntry struct {
	Index int `json:"i"`
	// Name     string            `json:"n"`
	Level       float64           `json:"lv"`
	StartOffset float64           `json:"so"`
	Duration    float64           `json:"du"`
	Layer       aggTraceLayer     `json:"ly,omitempty"`
	Data        map[string]string `json:"dt,omitempty"`
}

type aggTrace struct {
	aggTxnID
	StartTime int64                  `json:"start"`
	Duration  float64                `json:"duration"`
	R         *aggRequest            `json:"request"`
	Entries   []aggTraceEntry        `json:"entries"`
	Funcs     []string               `json:"funcs"`
	Partial   bool                   `json:"partial"`
	Custom    map[string]interface{} `json:"customData,omitempty"`
}

type aggTraceBatch struct {
	PrimarySet             aggTraceSet
	SecondarySet           aggTraceSet
	BackgroundPrimarySet   aggTraceSet
	BackgroundSecondarySet aggTraceSet
}

func (b *aggTraceBatch) Init() {
	b.PrimarySet.Init("primary", 5, true)
	b.SecondarySet.Init("secondary", 4, false)
	b.BackgroundPrimarySet.Init("backgroudPrimary", 5, true)
	b.BackgroundSecondarySet.Init("backgroundSecondary", 4, false)
}

func (b *aggTraceBatch) Add(key string, trace *aggTrace) {

	if trace.isBackgroundTxn() == false {
		index, ok := b.PrimarySet.Stat[key]
		if !ok {
			b.PrimarySet.Add(key, trace)
		} else {
			if b.PrimarySet.SavedCount < b.PrimarySet.AllowedCount {
				b.SecondarySet.Add(key, trace)
			} else {
				// if it is greater than 5, then we are going to send this set in all cases,
				// so no need to replace or check the secondary set
				if trace.Duration > b.PrimarySet.Traces[index].Duration {
					b.PrimarySet.Traces[index] = *trace
				}
			}
		}
	} else {
		index, ok := b.BackgroundPrimarySet.Stat[key]
		if !ok {
			b.BackgroundPrimarySet.Add(key, trace)
		} else {
			if b.BackgroundPrimarySet.SavedCount < b.BackgroundPrimarySet.AllowedCount {
				b.BackgroundSecondarySet.Add(key, trace)
			} else {
				// if it is greater than 5, then we are going to send this set in all cases,
				// so no need to replace or check the secondary set
				if trace.Duration > b.BackgroundPrimarySet.Traces[index].Duration {
					b.BackgroundPrimarySet.Traces[index] = *trace
				}
			}
		}
	}
}

type aggTraceSet struct {
	Name           string
	UniqueTraces   bool
	AllowedCount   int
	Stat           map[string]int
	SavedCount     int
	Traces         []aggTrace
	LowestDuration float64
	LowestIndex    int
}

// ByTraceDuration implements sort.Interface for []aggTrace based on
// the Duration field.
type ByTraceDuration []aggTrace

func (a ByTraceDuration) Len() int           { return len(a) }
func (a ByTraceDuration) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTraceDuration) Less(i, j int) bool { return a[i].Duration > a[j].Duration }

func (m *aggTraceSet) Init(name string, count int, uniqueTraces bool) {
	m.Name = name
	m.Traces = make([]aggTrace, 0, count)
	m.AllowedCount = count
	m.UniqueTraces = uniqueTraces
	if m.UniqueTraces {
		m.Stat = make(map[string]int, count)
	}
}

func (m *aggTraceSet) Add(key string, trace *aggTrace) {
	if m.SavedCount < m.AllowedCount {

		if m.SavedCount == 0 {
			m.LowestDuration = trace.Duration
			m.LowestIndex = 0
		}

		m.Traces = append(m.Traces, *trace)
		m.SavedCount++
		if m.UniqueTraces {
			m.Stat[key] = len(m.Traces) - 1
		}

		if trace.Duration < m.LowestDuration {
			m.LowestDuration = trace.Duration
			m.LowestIndex = len(m.Traces) - 1
		}

	} else {
		if trace.Duration > m.LowestDuration {
			if m.UniqueTraces {
				delete(m.Stat, m.Traces[m.LowestIndex].Key())
				m.Stat[key] = m.LowestIndex
			}

			m.Traces[m.LowestIndex] = *trace
			m.LowestDuration = trace.Duration
			for i := range m.Traces {
				if m.Traces[i].Duration < m.LowestDuration {
					m.LowestDuration = m.Traces[i].Duration
					m.LowestIndex = i
				}
			}
		}
	}
}

type tracePayload struct {
	header
	StartTime int64      `json:"startTime"`
	EndTime   int64      `json:"endTime"`
	T         []aggTrace `json:"traces"`
}
