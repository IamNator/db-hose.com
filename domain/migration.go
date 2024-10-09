package domain

import (
	"time"
)

type Migration struct {
	TimeStamp time.Time         `json:"timestamp" gorm:"not null"`
	Email     string            `json:"email" gorm:"not null"`
	Size      int64             `json:"size" gorm:"not null"`
	Meta      map[string]string `json:"meta"`
}

// File returns the file name of the migration
func (m Migration) File() string {
	return m.Email + "/" + m.TimeStamp.Format("2006_01_02_15_04_05") + ".sql"
}
