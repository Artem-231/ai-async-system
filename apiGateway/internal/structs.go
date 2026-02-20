package internal

import "time"

// Task тип, который предназначен для хранения данных в формате json
type Task struct {
	Id        string    `json:"id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

// Request тип, который предназначен для хранения данных в формате json
type Request struct {
	Priority int    `json:"priority"`
	Action   string `json:"action"`
	Payload  string `json:"payload"`
}
