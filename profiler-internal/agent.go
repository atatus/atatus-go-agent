// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2022 Datadog, Inc.

package internal

import (
	"net"
	"net/url"
	"os"

	"go.atatus.com/agent/profiler-internal/log"
)

const (
	DefaultAgentHostname  = "localhost"
	DefaultTraceAgentPort = "8091"
	defaultAPIURL         = "https://profiling-rx.atatus.com/track/profiling"
)

// This is a variable rather than a constant so it can be replaced in unit tests
var DefaultTraceAgentUDSPath = "/var/run/datadog/apm.socket"

// AgentURLFromEnv resolves the URL for the trace agent based on
// the default host/port and UDS path, and via standard environment variables.
// AgentURLFromEnv has the following priority order:
//   - First, DD_TRACE_AGENT_URL if it is set
//   - Then, if either of DD_AGENT_HOST and DD_TRACE_AGENT_PORT are set,
//     use http://DD_AGENT_HOST:DD_TRACE_AGENT_PORT,
//     defaulting to localhost and 8126, respectively
//   - Then, DefaultTraceAgentUDSPath, if the path exists
//   - Finally, localhost:8126
func AgentURLFromEnv() *url.URL {
	if agentURL := os.Getenv("ATATUS_PROFILING_URL"); agentURL != "" { // ATCHG - DD_TRACE_AGENT_URL changed into ATATUS_PROFILING_URL
		u, err := url.Parse(agentURL)
		if err != nil {
			log.Warn("Failed to parse ATATUS_PROFILING_URL: %s", err.Error())
		} else {
			switch u.Scheme {
			case "unix", "http", "https":
				return u
			default:
				log.Warn("Unsupported protocol %q in Agent URL %q. Must be one of: http, https, unix.", u.Scheme, agentURL)
			}
		}
	} else {
		u, err := url.Parse(defaultAPIURL)
		if err != nil {
			log.Warn("Failed to parse default API URL: %s", err.Error())
			return nil
		}
		return u
	}

	host, providedHost := os.LookupEnv("ATATUS_AGENT_HOST")       // ATCHG - DD_AGENT_HOST changed into ATATUS_AGENT_HOST
	port, providedPort := os.LookupEnv("ATATUS_TRACE_AGENT_PORT") //ATCHG - DD_TRACE_AGENT_PORT changed into ATATUS_TRACE_AGENT_PORT
	if host == "" {
		// We treat set but empty the same as unset
		providedHost = false
		host = DefaultAgentHostname
	}
	if port == "" {
		// We treat set but empty the same as unset
		providedPort = false
		port = DefaultTraceAgentPort
	}
	httpURL := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
	}
	if providedHost || providedPort {
		return httpURL
	}

	if _, err := os.Stat(DefaultTraceAgentUDSPath); err == nil {
		return &url.URL{
			Scheme: "unix",
			Path:   DefaultTraceAgentUDSPath,
		}
	}
	return httpURL
}
