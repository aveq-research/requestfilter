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
	FilterRegexes []string `json:"filterRegexes,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		FilterRegexes: []string{},
	}
}

// RequestFilter a request filtering plugin.
type RequestFilter struct {
	next          http.Handler
	name          string
	filterRegexes []*regexp.Regexp
}

// New creates a new RequestFilter plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	var regexes []*regexp.Regexp
	for _, pattern := range config.FilterRegexes {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid filter regex '%s': %v", pattern, err)
		}
		regexes = append(regexes, regex)
	}

	return &RequestFilter{
		next:          next,
		name:          name,
		filterRegexes: regexes,
	}, nil
}

func (r *RequestFilter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// os.Stdout.WriteString(fmt.Sprintf("RequestFilter: Processing request for URL: %s\n", req.URL.String()))

	// Check if we need to filter the request
	if len(r.filterRegexes) > 0 {
		// os.Stdout.WriteString(fmt.Sprintf("RequestFilter: Number of regexes: %d\n", len(r.filterRegexes)))

		// Check URL
		for _, regex := range r.filterRegexes {
			// os.Stdout.WriteString(fmt.Sprintf("RequestFilter: Checking regex %d: %s\n", i, regex.String()))
			if regex.MatchString(req.URL.String()) {
				// os.Stdout.WriteString(fmt.Sprintf("RequestFilter: URL matched regex %d. Blocking request.\n", i))
				http.Error(rw, "Request blocked by filter", http.StatusForbidden)
				return
			}
		}
		// os.Stdout.WriteString("RequestFilter: URL did not match any regexes\n")

		// Check body for POST and PUT requests
		if req.Method == http.MethodPost || req.Method == http.MethodPut {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				// os.Stdout.WriteString(fmt.Sprintf("RequestFilter: Error reading request body: %v\n", err))
				http.Error(rw, "Error reading request body", http.StatusInternalServerError)
				return
			}

			for _, regex := range r.filterRegexes {
				if regex.Match(body) {
					// os.Stdout.WriteString(fmt.Sprintf("RequestFilter: Body matched regex %d. Blocking request.\n", i))
					http.Error(rw, "Request blocked by filter", http.StatusForbidden)
					return
				}
			}
			// os.Stdout.WriteString("RequestFilter: Body did not match any regexes\n")

			// Replace the body with a new ReadCloser
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
	} else {
		// os.Stdout.WriteString("RequestFilter: No regexes configured\n")
	}

	// os.Stdout.WriteString("RequestFilter: Passing request to next handler\n")
	r.next.ServeHTTP(rw, req)
}
