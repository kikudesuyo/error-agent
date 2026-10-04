// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"time"

	"github.com/kikudesuyo/error-agent/api/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrorEventRecord is the PostgreSQL representation of an accepted event.
type ErrorEventRecord struct {
	Fingerprint string `gorm:"primaryKey;size:64"`
	Service     string `gorm:"size:64;not null"`
	Environment string `gorm:"size:64;not null"`
	Status      int    `gorm:"not null"`
	Method      string `gorm:"size:16;not null"`
	URL         string `gorm:"type:text;not null"`
	Path        string `gorm:"type:text;not null"`
	ErrorType   string `gorm:"size:128;not null"`
	Message     string `gorm:"type:text;not null"`
	Stack       string `gorm:"type:text;not null"`
	CommitSHA   string `gorm:"size:64;not null"`
	CFRay       string `gorm:"size:128;not null"`
	UserAgent   string `gorm:"size:1024;not null"`
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	Occurrences int64 `gorm:"not null"`
}

func (ErrorEventRecord) TableName() string { return "error_events" }

type ErrorEventPostgresRepository struct {
	db *gorm.DB
}

func NewErrorEventPostgresRepository(db *gorm.DB) *ErrorEventPostgresRepository {
	return &ErrorEventPostgresRepository{db: db}
}

// RecordErrorEvent inserts an event once and updates occurrence metadata on duplicates.
func (repository *ErrorEventPostgresRepository) RecordErrorEvent(ctx context.Context, fingerprint string, event entity.ErrorEvent) (bool, error) {
	now := time.Now().UTC()
	record := ErrorEventRecord{
		Fingerprint: fingerprint,
		Service:     event.Service,
		Environment: event.Environment,
		Status:      event.Status,
		Method:      event.Method,
		URL:         event.URL,
		Path:        event.Path,
		ErrorType:   event.ErrorType,
		Message:     event.Message,
		Stack:       event.Stack,
		CommitSHA:   event.CommitSHA,
		CFRay:       event.CFRay,
		UserAgent:   event.UserAgent,
		FirstSeenAt: now,
		LastSeenAt:  now,
		Occurrences: 1,
	}

	created := false
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 1 {
			created = true
			return nil
		}
		return tx.Model(&ErrorEventRecord{}).
			Where("fingerprint = ?", fingerprint).
			Updates(map[string]any{
				"last_seen_at": now,
				"occurrences":  gorm.Expr("occurrences + 1"),
			}).Error
	})
	return created, err
}
