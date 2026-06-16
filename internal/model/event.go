package model

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
