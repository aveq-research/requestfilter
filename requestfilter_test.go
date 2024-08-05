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

//nolint:gocognit
func TestRequestFilter(t *testing.T) {
	tests := []struct {
		name             string
		config           *requestfilter.Config
		method           string
		url              string
		body             string
		expectedStatus   int
		expectedBody     string
		expectedResponse string
	}{
		{
			name: "No filter",
			config: &requestfilter.Config{
				FilterRegexes: []string{},
			},
			method:           http.MethodGet,
			url:              "http://example.com",
			expectedStatus:   http.StatusOK,
			expectedResponse: "",
		},
		{
			name: "URL path filter - blocked",
			config: &requestfilter.Config{
				FilterRegexes: []string{"sensitive"},
			},
			method:         http.MethodGet,
			url:            "http://example.com/sensitive-data",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked\n",
		},
		{
			name: "Body filter - POST blocked",
			config: &requestfilter.Config{
				FilterRegexes: []string{"confidential"},
			},
			method:         http.MethodPost,
			url:            "http://example.com",
			body:           "This contains confidential information",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked\n",
		},
		{
			name: "Body filter - PATCH blocked",
			config: &requestfilter.Config{
				FilterRegexes: []string{"classified"},
			},
			method:         http.MethodPatch,
			url:            "http://example.com",
			body:           "This is classified information",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Request blocked\n",
		},
		{
			name: "Custom error message",
			config: &requestfilter.Config{
				FilterRegexes:    []string{"sensitive"},
				HTTPErrorMessage: "Access Denied",
			},
			method:         http.MethodGet,
			url:            "http://example.com/sensitive-data",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Access Denied\n",
		},
		{
			name: "PathOnly - body not checked",
			config: &requestfilter.Config{
				FilterRegexes: []string{"confidential"},
				PathOnly:      true,
			},
			method:           http.MethodPost,
			url:              "http://example.com",
			body:             "This contains confidential information",
			expectedStatus:   http.StatusOK,
			expectedResponse: "This contains confidential information",
		},
		{
			name: "BodyOnly - path not checked",
			config: &requestfilter.Config{
				FilterRegexes: []string{"sensitive"},
				BodyOnly:      true,
			},
			method:         http.MethodGet,
			url:            "http://example.com/sensitive-data",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.Body != nil {
					body, _ := io.ReadAll(req.Body)
					_, err := rw.Write(body)
					if err != nil {
						t.Fatal(err)
					}
				}
			})

			handler, err := requestfilter.New(ctx, next, tt.config, "request-filter")
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

			if tt.expectedResponse != "" && recorder.Body.String() != tt.expectedResponse {
				t.Errorf("Expected response '%s', got '%s'", tt.expectedResponse, recorder.Body.String())
			}
		})
	}
}

func TestCreateConfig(t *testing.T) {
	cfg := requestfilter.CreateConfig()
	if cfg == nil {
		t.Error("CreateConfig returned nil")
	}
	if cfg != nil && len(cfg.FilterRegexes) != 0 {
		t.Errorf("Expected empty FilterRegexes, got %v", cfg.FilterRegexes)
	}
	if cfg.HTTPErrorMessage != "" {
		t.Errorf("Expected empty HttpErrorMessage, got %v", cfg.HTTPErrorMessage)
	}
	if cfg.PathOnly != false {
		t.Errorf("Expected PathOnly to be false, got %v", cfg.PathOnly)
	}
	if cfg.BodyOnly != false {
		t.Errorf("Expected BodyOnly to be false, got %v", cfg.BodyOnly)
	}
}

func TestNewWithInvalidRegex(t *testing.T) {
	cfg := requestfilter.CreateConfig()
	cfg.FilterRegexes = []string{"["} // Invalid regex

	ctx := context.Background()
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	_, err := requestfilter.New(ctx, next, cfg, "request-filter")
	if err == nil {
		t.Error("Expected error for invalid regex, got nil")
	}
}
