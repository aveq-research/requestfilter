// Package requestfilter provides a Traefik middleware for filtering HTTP requests.
package requestfilter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// Config the plugin configuration.
type Config struct {
	FilterRegexes    []string `json:"filterRegexes,omitempty"`
	HTTPErrorMessage string   `json:"httpErrorMessage,omitempty"`
	PathOnly         bool     `json:"pathOnly,omitempty"`
	BodyOnly         bool     `json:"bodyOnly,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		FilterRegexes:    []string{},
		HTTPErrorMessage: "",
		PathOnly:         false,
		BodyOnly:         false,
	}
}

// RequestFilter a request filtering plugin.
type RequestFilter struct {
	next             http.Handler
	name             string
	filterRegexes    []*regexp.Regexp
	httpErrorMessage string
	pathOnly         bool
	bodyOnly         bool
}

// New creates a new RequestFilter plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	regexes := make([]*regexp.Regexp, 0, len(config.FilterRegexes))
	for _, pattern := range config.FilterRegexes {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid filter regex '%s': %v", pattern, err)
		}
		regexes = append(regexes, regex)
	}

	// validate that at only one of pathOnly and bodyOnly is set
	if config.PathOnly && config.BodyOnly {
		//nolint:perfsprint
		return nil, fmt.Errorf("only one of pathOnly and bodyOnly can be set")
	}

	return &RequestFilter{
		next:             next,
		name:             name,
		filterRegexes:    regexes,
		httpErrorMessage: config.HTTPErrorMessage,
		pathOnly:         config.PathOnly,
		bodyOnly:         config.BodyOnly,
	}, nil
}

func (r *RequestFilter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	//nolint:nestif
	if len(r.filterRegexes) > 0 {
		// Check URL path if not bodyOnly
		if !r.bodyOnly {
			for _, regex := range r.filterRegexes {
				if regex.MatchString(req.URL.Path) {
					r.block(rw)
					return
				}
			}
		}

		// Check body for POST, PUT, and PATCH requests if not pathOnly
		if !r.pathOnly && (req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(rw, "Error reading request body", http.StatusInternalServerError)
				return
			}

			for _, regex := range r.filterRegexes {
				if regex.Match(body) {
					r.block(rw)
					return
				}
			}

			// Replace the body with a new ReadCloser
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
	}

	r.next.ServeHTTP(rw, req)
}

func (r *RequestFilter) block(rw http.ResponseWriter) {
	if r.httpErrorMessage == "" {
		http.Error(rw, "Request blocked", http.StatusForbidden)
	} else {
		http.Error(rw, r.httpErrorMessage, http.StatusForbidden)
	}
}
