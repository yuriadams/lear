package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Book struct {
	ID          int
	GutenbergID int
	Content     string
	Metadata    Metadata
	CreatedAt   time.Time
	DeletedAt   time.Time
}

type Metadata struct {
	Author   string `json:"author"`
	Title    string `json:"title"`
	Credits  string `json:"credits"`
	Summary  string `json:"summary"`
	Language string `json:"language"`
	Subject  string `json:"subject"`
	Category string `json:"category"`
}

func (m Metadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Metadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &m)
}
