package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	LeadStatusPending    = "pending"
	LeadStatusCalling    = "calling"
	LeadStatusCompleted  = "completed"
	LeadStatusInterested = "interested"
	LeadStatusFailed     = "failed"
)

type Lead struct {
	ID uuid.UUID `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey"`

	FullName    string
	Company     string
	PhoneNumber string

	Status string
	Error  *string

	GoogleSheetSynced bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Lead) TableName() string {
	return "leads"
}
