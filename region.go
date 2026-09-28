package atatus // import "go.atatus.com/agent"

import (
	"net"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

var validRegions = []string{"in", "eu"}

func isValidRegion(region string) bool {
	for _, r := range validRegions {
		if r == region {
			return true
		}
	}
	return false
}

// applyRegionToURL inserts "-<region>" before the first dot of rawURL's host,
// preserving the scheme, port and path. A host without a dot is returned as is.
func applyRegionToURL(rawURL, region string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", errors.Wrapf(err, "invalid server URL %q", rawURL)
	}

	host := u.Hostname()
	dot := strings.Index(host, ".")
	if dot < 0 {
		return rawURL, nil
	}

	host = host[:dot] + "-" + region + host[dot:]
	if port := u.Port(); port != "" {
		host = net.JoinHostPort(host, port)
	}
	u.Host = host

	return u.String(), nil
}

// applyRegion rewrites the default server URLs to the configured Atatus data
// region. It must be called after all server URLs have been resolved. When no
// region is configured the server URLs are left untouched.
//
// The region is mutually exclusive with an explicitly configured server URL:
// either the region decides where data is sent, or the user does.
func (opts *TracerOptions) applyRegion() error {
	region := opts.Region
	if region == "" {
		region = initialRegion()
	}
	region = strings.ToLower(strings.TrimSpace(region))

	if region == "" {
		return nil
	}

	if !isValidRegion(region) {
		return errors.Errorf("invalid %s %q. Agent disabled", envRegion, region)
	}

	conflicts := []struct {
		name  string
		isSet bool
	}{
		{envServiceNotifyHost, opts.NotifyHost != ""},
		{envServerUrl, opts.ServerUrl != ""},
		{envApmServerUrl, opts.ApmServerUrl != defaultApmServerUrl},
		{envAnalyticsServerUrl, opts.AnalyticsServerUrl != defaultAnalyticsServerUrl},
		{envTracesServerUrl, opts.TracesServerUrl != defaultTracesServerUrl},
		{envLogsServerUrl, opts.LogsServerUrl != defaultLogsServerUrl},
		{envProfilingServerUrl, opts.ProfilingServerUrl != defaultProfilingServerUrl},
	}
	for _, c := range conflicts {
		if c.isSet {
			return errors.Errorf(
				"%s cannot be used with %s. Remove one of them. Agent disabled",
				envRegion, c.name,
			)
		}
	}

	targets := []*string{
		&opts.ApmServerUrl,
		&opts.AnalyticsServerUrl,
		&opts.TracesServerUrl,
		&opts.LogsServerUrl,
		&opts.ProfilingServerUrl,
	}
	for _, target := range targets {
		regionURL, err := applyRegionToURL(*target, region)
		if err != nil {
			return err
		}
		*target = regionURL
	}

	return nil
}
