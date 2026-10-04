// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	"github.com/kikudesuyo/error-agent/api/entity"
)

var (
	cookiePattern = regexp.MustCompile(`(?i)\b(cookie|set-cookie)\b\s*[:=][^\r\n]*`)
	secretPattern = regexp.MustCompile(`(?i)\b(authorization|proxy-authorization|password|passwd|secret|client[_-]?secret|access[_-]?token|refresh[_-]?token|token|api[_-]?key|private[_-]?key)\b\s*[:=]\s*([^\s,;]+)`)
	bearerPattern = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+`)
	emailPattern  = regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
	phonePattern  = regexp.MustCompile(`\+?\d[\d .()-]{8,}\d`)
)

// ErrorEventStore persists an error event once per fingerprint.
type ErrorEventStore interface {
	RecordErrorEvent(ctx context.Context, fingerprint string, event entity.ErrorEvent) (bool, error)
}

type ErrorEventService struct {
	store ErrorEventStore
}

func NewErrorEventService(store ErrorEventStore) *ErrorEventService {
	return &ErrorEventService{store: store}
}

// AcceptErrorEvent redacts likely secrets and personal identifiers before persistence.
func (service *ErrorEventService) AcceptErrorEvent(ctx context.Context, request entity.PostErrorEventRequest) (entity.PostErrorEventResponse, error) {
	event := redactEvent(request)
	fingerprint := createFingerprint(event)
	created, err := service.store.RecordErrorEvent(ctx, fingerprint, event)
	if err != nil {
		return entity.PostErrorEventResponse{}, err
	}
	return entity.PostErrorEventResponse{Fingerprint: fingerprint, Duplicate: !created}, nil
}

func redactEvent(request entity.PostErrorEventRequest) entity.ErrorEvent {
	parsedURL := request.URL
	if parsedURL != "" {
		parsedURL = redactText(parsedURL)
	}
	return entity.ErrorEvent{
		Service:     request.Service,
		Environment: request.Environment,
		Status:      request.Status,
		Method:      request.Method,
		URL:         parsedURL,
		Path:        redactText(request.Path),
		ErrorType:   redactText(request.ErrorType),
		Message:     redactText(request.Message),
		Stack:       redactText(request.Stack),
		CommitSHA:   request.CommitSHA,
		CFRay:       request.CFRay,
		UserAgent:   redactText(request.UserAgent),
	}
}

func redactText(value string) string {
	value = cookiePattern.ReplaceAllString(value, "$1=[REDACTED]")
	value = bearerPattern.ReplaceAllString(value, "Bearer [REDACTED]")
	value = secretPattern.ReplaceAllString(value, "$1=[REDACTED]")
	value = emailPattern.ReplaceAllString(value, "[REDACTED_EMAIL]")
	return phonePattern.ReplaceAllString(value, "[REDACTED_PHONE]")
}

func createFingerprint(event entity.ErrorEvent) string {
	stackFirstLine, _, _ := strings.Cut(event.Stack, "\n")
	value := strings.Join([]string{
		event.Service,
		event.Environment,
		strconv.Itoa(event.Status),
		event.Method,
		event.Path,
		event.ErrorType,
		event.Message,
		stackFirstLine,
	}, "\x00")
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
