package requestfilter_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aveq-research/requestfilter"
)

func TestRequestFilter(t *testing.T) {
	tests := []struct {
		name           string
		filterRegexes  []string
		method         string
		url            string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No filter",
			filterRegexes:  []string{},
			method:         http.MethodGet,
			url:            "http://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Single URL filter - blocked",
			filterRegexes:  []string{"sensitive"},
			method:         http.MethodGet,
			url:            "http://example.com/sensitive-data",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked by filter\n",
		},
		{
			name:           "Single URL filter - allowed",
			filterRegexes:  []string{"sensitive"},
			method:         http.MethodGet,
			url:            "http://example.com/public-data",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Multiple URL filters - blocked",
			filterRegexes:  []string{"sensitive", "private", "confidential"},
			method:         http.MethodGet,
			url:            "http://example.com/private-area",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked by filter\n",
		},
		{
			name:           "Multiple URL filters - allowed",
			filterRegexes:  []string{"sensitive", "private", "confidential"},
			method:         http.MethodGet,
			url:            "http://example.com/public-area",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Single body filter - POST blocked",
			filterRegexes:  []string{"confidential"},
			method:         http.MethodPost,
			url:            "http://example.com",
			body:           "This contains confidential information",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked by filter\n",
		},
		{
			name:           "Single body filter - POST allowed",
			filterRegexes:  []string{"confidential"},
			method:         http.MethodPost,
			url:            "http://example.com",
			body:           "This contains public information",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Multiple body filters - PUT blocked",
			filterRegexes:  []string{"secret", "classified", "confidential"},
			method:         http.MethodPut,
			url:            "http://example.com",
			body:           "This is a classified message",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked by filter\n",
		},
		{
			name:           "Multiple body filters - PUT allowed",
			filterRegexes:  []string{"secret", "classified", "confidential"},
			method:         http.MethodPut,
			url:            "http://example.com",
			body:           "This is a public message",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Complex regex - email blocked",
			filterRegexes:  []string{`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`},
			method:         http.MethodPost,
			url:            "http://example.com",
			body:           "Please contact me at user@example.com",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked by filter\n",
		},
		{
			name:           "Complex regex - phone number blocked",
			filterRegexes:  []string{`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`},
			method:         http.MethodPost,
			url:            "http://example.com",
			body:           "Call me at 123-456-7890",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked by filter\n",
		},
		{
			name:           "Multiple complex regexes - allowed",
			filterRegexes:  []string{`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`, `\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`},
			method:         http.MethodPost,
			url:            "http://example.com",
			body:           "This is a public message without sensitive information",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := requestfilter.CreateConfig()
			cfg.FilterRegexes = tt.filterRegexes

			ctx := context.Background()
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Echo back the request body for body preservation test
				if req.Body != nil {
					body, _ := io.ReadAll(req.Body)
					rw.Write(body)
				}
			})

			handler, err := requestfilter.New(ctx, next, cfg, "request-filter")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()

			var req *http.Request
			if tt.body != "" {
				req, err = http.NewRequestWithContext(ctx, tt.method, tt.url, strings.NewReader(tt.body))
			} else {
				req, err = http.NewRequestWithContext(ctx, tt.method, tt.url, nil)
			}
			if err != nil {
				t.Fatal(err)
			}

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedBody != "" && recorder.Body.String() != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, recorder.Body.String())
			}

			if tt.expectedStatus == http.StatusOK && tt.body != "" {
				// Check if the body is preserved for allowed requests
				if recorder.Body.String() != tt.body {
					t.Errorf("Body not preserved. Expected '%s', got '%s'", tt.body, recorder.Body.String())
				}
			}
		})
	}
}

func TestCreateConfig(t *testing.T) {
	cfg := requestfilter.CreateConfig()
	if cfg == nil {
		t.Error("CreateConfig returned nil")
	}
	if len(cfg.FilterRegexes) != 0 {
		t.Errorf("Expected empty FilterRegexes, got %v", cfg.FilterRegexes)
	}
}

func TestNewWithInvalidRegex(t *testing.T) {
	cfg := requestfilter.CreateConfig()
	cfg.FilterRegexes = []string{"["} // Invalid regex

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := requestfilter.New(ctx, next, cfg, "request-filter")
	if err == nil {
		t.Error("Expected error for invalid regex, got nil")
	}
}
