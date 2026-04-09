package model

type JobRequest struct {
	URL         string
	Method      string
	Header      map[string]string
	Body        []byte
	Concurrency int
}

type RunRequest struct {
	URL         string `json:"url"`
	Method      string `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        string `json:"body"`
	Concurrency int    `json:"concurrency"`
}
