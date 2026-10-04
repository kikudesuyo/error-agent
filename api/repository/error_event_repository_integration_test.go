// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/kikudesuyo/error-agent/api/entity"
	"github.com/kikudesuyo/error-agent/api/library"
)

func TestRecordErrorEventDeduplicates(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := library.OpenDatabase(databaseURL)
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	testFingerprintData := []byte(time.Now().UTC().Format(time.RFC3339Nano))
	testFingerprintHash := sha256.Sum256(testFingerprintData)
	fingerprint := hex.EncodeToString(testFingerprintHash[:])
	cleanup := db.Exec("DELETE FROM error_events WHERE fingerprint = ?", fingerprint)
	if cleanup.Error != nil {
		t.Fatalf("clean test event before insert: %v", cleanup.Error)
	}
	t.Cleanup(func() { _ = db.Exec("DELETE FROM error_events WHERE fingerprint = ?", fingerprint).Error })

	event := entity.ErrorEvent{
		Service: "buildlog-frontend", Environment: "production", Status: 500,
		Method: "GET", URL: "https://example.com/contact", Path: "/contact",
		ErrorType: "Error", Message: "request failed", Stack: "Error: request failed",
	}
	repository := NewErrorEventPostgresRepository(db)
	created, err := repository.RecordErrorEvent(context.Background(), fingerprint, event)
	if err != nil {
		t.Fatalf("first RecordErrorEvent() error = %v", err)
	}
	if !created {
		t.Fatal("first RecordErrorEvent() created = false, want true")
	}

	duplicateEvent := event
	duplicateEvent.Message = "new message should not replace the first event"
	created, err = repository.RecordErrorEvent(context.Background(), fingerprint, duplicateEvent)
	if err != nil {
		t.Fatalf("duplicate RecordErrorEvent() error = %v", err)
	}
	if created {
		t.Fatal("duplicate RecordErrorEvent() created = true, want false")
	}

	var storedRecord ErrorEventRecord
	if err := db.First(&storedRecord, "fingerprint = ?", fingerprint).Error; err != nil {
		t.Fatalf("read stored event: %v", err)
	}
	if storedRecord.Occurrences != 2 {
		t.Errorf("occurrences = %d, want 2", storedRecord.Occurrences)
	}
	if storedRecord.Message != event.Message {
		t.Errorf("duplicate replaced first message: got %q, want %q", storedRecord.Message, event.Message)
	}
}
