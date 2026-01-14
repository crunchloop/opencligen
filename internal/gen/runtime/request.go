package runtime

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// pathParamRegex matches path parameters like {id} or {userId}
var pathParamRegex = regexp.MustCompile(`\{([^}]+)\}`)

// Request represents an HTTP request to be executed
type Request struct {
	Method      string
	Path        string
	PathParams  map[string]string
	QueryParams map[string]string
	Headers     map[string]string
	Body        []byte
}

// NewRequest creates a new Request
func NewRequest(method, path string) *Request {
	return &Request{
		Method:      method,
		Path:        path,
		PathParams:  make(map[string]string),
		QueryParams: make(map[string]string),
		Headers:     make(map[string]string),
	}
}

// SetPathParam sets a path parameter
func (r *Request) SetPathParam(name, value string) {
	r.PathParams[name] = value
}

// SetQueryParam sets a query parameter
func (r *Request) SetQueryParam(name, value string) {
	r.QueryParams[name] = value
}

// SetHeader sets a header
func (r *Request) SetHeader(name, value string) {
	r.Headers[name] = value
}

// SetBody sets the request body
func (r *Request) SetBody(body []byte) {
	r.Body = body
}

// Build creates an http.Request from this Request
func (r *Request) Build(ctx context.Context, baseURL string) (*http.Request, error) {
	// Validate all path parameters are provided
	matches := pathParamRegex.FindAllStringSubmatch(r.Path, -1)
	var missing []string
	for _, match := range matches {
		paramName := match[1]
		if _, ok := r.PathParams[paramName]; !ok {
			missing = append(missing, paramName)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required path parameter(s): %s", strings.Join(missing, ", "))
	}

	// Substitute path parameters
	path := r.Path
	for name, value := range r.PathParams {
		placeholder := "{" + name + "}"
		path = strings.ReplaceAll(path, placeholder, url.PathEscape(value))
	}

	// Build full URL
	fullURL := strings.TrimSuffix(baseURL, "/") + path

	// Add query parameters
	if len(r.QueryParams) > 0 {
		params := url.Values{}
		for name, value := range r.QueryParams {
			params.Add(name, value)
		}
		fullURL += "?" + params.Encode()
	}

	// Create request
	var bodyReader *bytes.Reader
	if r.Body != nil {
		bodyReader = bytes.NewReader(r.Body)
	}

	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequestWithContext(ctx, r.Method, fullURL, bodyReader)
	} else {
		req, err = http.NewRequestWithContext(ctx, r.Method, fullURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for name, value := range r.Headers {
		req.Header.Set(name, value)
	}

	// Set content-type for JSON body
	if r.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
