package model

type RequestJob struct {
	URL         string
	Method      string
	Header      map[string]string
	Body        []byte
	Concurrency int
}
