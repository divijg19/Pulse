package model

type Event struct {
	Type  string `json:"type"`
	RunID string `json:"run_id"`
	Data  any    `json:"data"`
}
