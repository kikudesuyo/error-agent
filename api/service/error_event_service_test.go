// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kikudesuyo/error-agent/api/entity"
)

type errorEventStoreStub struct {
	created     bool
	storedEvent entity.ErrorEvent
	fingerprint string
	err         error
}

func (store *errorEventStoreStub) RecordErrorEvent(_ context.Context, fingerprint string, event entity.ErrorEvent) (bool, error) {
	store.fingerprint = fingerprint
	store.storedEvent = event
	return store.created, store.err
}

func TestAcceptErrorEventRedactsSensitiveText(t *testing.T) {
	store := &errorEventStoreStub{created: true}
	service := NewErrorEventService(store)
	request := entity.PostErrorEventRequest{
		Service: "buildlog-frontend", Environment: "production", Status: 500, Method: "GET",
		URL: "https://example.com/contact", Path: "/contact",
		ErrorType: "Error", Message: "Bearer ghp_exampletoken password=hunter2 contact person@example.com 090-1234-5678",
		Stack: "Error: authorization: Bearer secret-token\nCookie: session=session-secret; theme=dark\n at handler", UserAgent: "person@example.com",
	}

	response, err := service.AcceptErrorEvent(context.Background(), request)
	if err != nil {
		t.Fatalf("AcceptErrorEvent() error = %v", err)
	}
	if response.Duplicate {
		t.Fatal("Duplicate = true, want false")
	}
	if len(response.Fingerprint) != 64 {
		t.Fatalf("fingerprint length = %d, want 64", len(response.Fingerprint))
	}
	storedData := strings.Join([]string{store.storedEvent.URL, store.storedEvent.Path, store.storedEvent.Message, store.storedEvent.Stack, store.storedEvent.UserAgent}, "\n")
	for _, sensitiveValue := range []string{"ghp_exampletoken", "hunter2", "person@example.com", "secret-token", "090-1234-5678", "session-secret", "theme=dark"} {
		if strings.Contains(storedData, sensitiveValue) {
			t.Errorf("stored event contains sensitive value %q", sensitiveValue)
		}
	}
}

func TestAcceptErrorEventReturnsStoreError(t *testing.T) {
	store := &errorEventStoreStub{err: errors.New("database unavailable")}
	service := NewErrorEventService(store)
	request := entity.PostErrorEventRequest{Service: "buildlog-frontend", Environment: "production", Status: 500, Method: "GET", URL: "https://example.com/", Path: "/"}
	if _, err := service.AcceptErrorEvent(context.Background(), request); err == nil {
		t.Fatal("AcceptErrorEvent() error = nil, want store error")
	}
}
