// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package atatus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRegionToURL(t *testing.T) {
	for _, tc := range []struct {
		name     string
		url      string
		region   string
		expected string
	}{
		{"apm", defaultApmServerUrl, "in", "https://apm-rx-in.atatus.com"},
		{"analytics", defaultAnalyticsServerUrl, "in", "https://an-rx-in.atatus.com"},
		{"traces", defaultTracesServerUrl, "in", "https://dt-rx-in.atatus.com"},
		{"logs", defaultLogsServerUrl, "in", "https://log-rx-in.atatus.com"},
		{"profiling", defaultProfilingServerUrl, "in", "https://profiling-rx-in.atatus.com"},
		{"eu", defaultApmServerUrl, "eu", "https://apm-rx-eu.atatus.com"},
		{"port", "https://apm-rx.atatus.com:8200", "in", "https://apm-rx-in.atatus.com:8200"},
		{"path", "https://apm-rx.atatus.com/track/apm", "in", "https://apm-rx-in.atatus.com/track/apm"},
		{"http", "http://apm-rx.atatus.com", "in", "http://apm-rx-in.atatus.com"},
		{"no dot", "http://localhost:8200", "in", "http://localhost:8200"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := applyRegionToURL(tc.url, tc.region)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestApplyRegionToURLInvalid(t *testing.T) {
	_, err := applyRegionToURL("://nope", "in")
	assert.Error(t, err)
}

func defaultURLOptions() TracerOptions {
	return TracerOptions{
		ApmServerUrl:       defaultApmServerUrl,
		AnalyticsServerUrl: defaultAnalyticsServerUrl,
		TracesServerUrl:    defaultTracesServerUrl,
		LogsServerUrl:      defaultLogsServerUrl,
		ProfilingServerUrl: defaultProfilingServerUrl,
	}
}

func TestApplyRegionUnsetLeavesDefaults(t *testing.T) {
	for _, region := range []string{"", "  "} {
		t.Run(region, func(t *testing.T) {
			opts := defaultURLOptions()
			opts.Region = region
			require.NoError(t, opts.applyRegion())

			assert.Equal(t, defaultApmServerUrl, opts.ApmServerUrl)
			assert.Equal(t, defaultAnalyticsServerUrl, opts.AnalyticsServerUrl)
			assert.Equal(t, defaultTracesServerUrl, opts.TracesServerUrl)
			assert.Equal(t, defaultLogsServerUrl, opts.LogsServerUrl)
			assert.Equal(t, defaultProfilingServerUrl, opts.ProfilingServerUrl)
		})
	}
}

func TestApplyRegionRewritesAll(t *testing.T) {
	for _, region := range []string{"in", "IN", " in "} {
		t.Run(region, func(t *testing.T) {
			opts := defaultURLOptions()
			opts.Region = region
			require.NoError(t, opts.applyRegion())

			assert.Equal(t, "https://apm-rx-in.atatus.com", opts.ApmServerUrl)
			assert.Equal(t, "https://an-rx-in.atatus.com", opts.AnalyticsServerUrl)
			assert.Equal(t, "https://dt-rx-in.atatus.com", opts.TracesServerUrl)
			assert.Equal(t, "https://log-rx-in.atatus.com", opts.LogsServerUrl)
			assert.Equal(t, "https://profiling-rx-in.atatus.com", opts.ProfilingServerUrl)
		})
	}
}

func TestApplyRegionInvalid(t *testing.T) {
	opts := defaultURLOptions()
	opts.Region = "xyz"

	err := opts.applyRegion()
	require.Error(t, err)
	assert.Contains(t, err.Error(), envRegion)
	assert.Contains(t, err.Error(), "xyz")
	assert.Equal(t, defaultApmServerUrl, opts.ApmServerUrl)
}

func TestApplyRegionConflicts(t *testing.T) {
	for _, tc := range []struct {
		envVar string
		mutate func(*TracerOptions)
	}{
		{envServiceNotifyHost, func(o *TracerOptions) { o.NotifyHost = "https://apm.corp.internal" }},
		{envServerUrl, func(o *TracerOptions) { o.ServerUrl = "https://apm.corp.internal" }},
		{envApmServerUrl, func(o *TracerOptions) { o.ApmServerUrl = "https://apm.corp.internal" }},
		{envAnalyticsServerUrl, func(o *TracerOptions) { o.AnalyticsServerUrl = "https://an.corp.internal" }},
		{envTracesServerUrl, func(o *TracerOptions) { o.TracesServerUrl = "https://dt.corp.internal" }},
		{envLogsServerUrl, func(o *TracerOptions) { o.LogsServerUrl = "https://logs.corp.internal" }},
		{envProfilingServerUrl, func(o *TracerOptions) { o.ProfilingServerUrl = "https://prof.corp.internal" }},
	} {
		t.Run(tc.envVar, func(t *testing.T) {
			opts := defaultURLOptions()
			opts.Region = "in"
			tc.mutate(&opts)

			err := opts.applyRegion()
			require.Error(t, err)
			assert.Contains(t, err.Error(), envRegion)
			assert.Contains(t, err.Error(), tc.envVar)
		})
	}
}

func TestApplyRegionFromEnv(t *testing.T) {
	t.Setenv(envRegion, "eu")

	opts := defaultURLOptions()
	require.NoError(t, opts.applyRegion())
	assert.Equal(t, "https://apm-rx-eu.atatus.com", opts.ApmServerUrl)
}

func TestInitDefaultsInvalidRegionDisablesTracer(t *testing.T) {
	t.Setenv(envRegion, "xyz")

	var opts TracerOptions
	require.NoError(t, opts.initDefaults(true))
	assert.False(t, opts.active)

	tracer := newTracer(opts)
	defer tracer.Close()
	assert.False(t, tracer.Active())
}

func TestInitDefaultsRegionConflictDisablesTracer(t *testing.T) {
	t.Setenv(envRegion, "in")
	t.Setenv(envServerUrl, "https://apm.corp.internal")

	var opts TracerOptions
	require.NoError(t, opts.initDefaults(true))
	assert.False(t, opts.active)
}

func TestNewTracerOptionsInvalidRegionReturnsError(t *testing.T) {
	t.Setenv(envRegion, "xyz")

	tracer, err := NewTracerOptions(TracerOptions{ServiceName: "region-test"})
	require.Error(t, err)
	assert.Nil(t, tracer)
}
