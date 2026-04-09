package model

import "time"

type Result struct {
	ID        string
	Status    int
	Latency   time.Duration
	Timestamp time.Time
	Error     string
	ResponseHeaders map[string]string
	ResponseBody    string
}
