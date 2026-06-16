package runconfig

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/divijg19/Pulse/internal/model"
)

const (
	DefaultMethod      = http.MethodGet
	DefaultConcurrency = 10
	MinConcurrency     = 1
	MaxConcurrency     = 100
)

var allowedMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
}

var allowedMethodsJoined = strings.Join(allowedMethods, ", ")

func AllowedMethods() []string {
	result := make([]string, len(allowedMethods))
	copy(result, allowedMethods)
	return result
}

func Validate(req model.RunRequest) (model.RunRequest, error) {
	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		return model.RunRequest{}, errors.New("URL is required")
	}

	parsed, err := url.ParseRequestURI(req.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return model.RunRequest{}, errors.New("URL must be a valid absolute HTTP URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return model.RunRequest{}, errors.New("URL scheme must be http or https")
	}

	req.Method = strings.ToUpper(strings.TrimSpace(req.Method))
	if req.Method == "" {
		req.Method = DefaultMethod
	}
	if !methodAllowed(req.Method) {
		return model.RunRequest{}, fmt.Errorf("method must be one of: %s", allowedMethodsJoined)
	}

	if req.Concurrency < MinConcurrency || req.Concurrency > MaxConcurrency {
		return model.RunRequest{}, fmt.Errorf("concurrency must be between %d and %d", MinConcurrency, MaxConcurrency)
	}

	if req.Headers == nil {
		req.Headers = map[string]string{}
	}

	return req, nil
}

func ClampConcurrency(value int) int {
	if value < MinConcurrency {
		return MinConcurrency
	}
	if value > MaxConcurrency {
		return MaxConcurrency
	}
	return value
}

func methodAllowed(method string) bool {
	for _, allowed := range allowedMethods {
		if method == allowed {
			return true
		}
	}
	return false
}
