package model

import "time"

type Result struct {
	Status          int               `json:"status"`
	Latency         time.Duration     `json:"latencyNs"`
	Timestamp       time.Time         `json:"timestamp"`
	Error           string            `json:"error,omitempty"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    string            `json:"responseBody,omitempty"`
	RequestMethod   string            `json:"requestMethod,omitempty"`
	RequestURL      string            `json:"requestUrl,omitempty"`
}
