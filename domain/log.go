package domain

import (
	"encoding/json"
	"time"
)

type Log struct {
	Timestamp   string `json:"timestamp"` // 2006-01-02 15:04:05
	Duration    string `json:"duration"`  // 1h2m3s
	Email       string `json:"email"`
	Name        string `json:"name"`
	Description string `json:"description"`
	File        string `json:"file"`
}

func NewLog(timestamp time.Time, duration, email, name, description, file string) Log {
	return Log{
		Timestamp:   timestamp.Format("2006-01-02 15:04:05"),
		Duration:    duration,
		Email:       email,
		Name:        name,
		Description: description,
		File:        file,
	}
}

func (l Log) Marshal() ([]byte, error) {
	return json.Marshal(l)
}
