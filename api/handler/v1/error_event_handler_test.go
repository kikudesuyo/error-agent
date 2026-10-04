// SPDX-License-Identifier: Apache-2.0
package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kikudesuyo/error-agent/api/entity"
)

type errorEventAcceptorStub struct {
	response entity.PostErrorEventResponse
	err      error
}

func (service errorEventAcceptorStub) AcceptErrorEvent(_ context.Context, _ entity.PostErrorEventRequest) (entity.PostErrorEventResponse, error) {
	return service.response, service.err
}

func TestHandlePostErrorEvent(t *testing.T) {
	testCaseList := []struct {
		name        string
		contentType string
		body        string
		service     errorEventAcceptorStub
		wantStatus  int
		wantBody    string
	}{
		{
			name: "accepts valid event", contentType: "application/json; charset=utf-8",
			body:       `{"service":"buildlog-frontend","environment":"production","status":500,"method":"GET","url":"https://example.com/contact","path":"/contact"}`,
			service:    errorEventAcceptorStub{response: entity.PostErrorEventResponse{Fingerprint: strings.Repeat("a", 64)}},
			wantStatus: http.StatusAccepted, wantBody: `"duplicate":false`,
		},
		{
			name: "rejects unknown fields", contentType: "application/json",
			body:       `{"service":"buildlog-frontend","unexpected":"value"}`,
			wantStatus: http.StatusBadRequest, wantBody: "invalid event",
		},
		{
			name: "rejects non-json content type", contentType: "text/plain", body: `{}`,
			wantStatus: http.StatusUnsupportedMediaType, wantBody: "content type",
		},
		{
			name: "does not expose storage error", contentType: "application/json",
			body:       `{"service":"buildlog-frontend","environment":"production","status":500,"method":"GET","url":"https://example.com/contact","path":"/contact"}`,
			service:    errorEventAcceptorStub{err: errors.New("secret database detail")},
			wantStatus: http.StatusServiceUnavailable, wantBody: "event could not be accepted",
		},
	}

	for _, testCase := range testCaseList {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewErrorEventHandler(testCase.service)
			request := httptest.NewRequest(http.MethodPost, "/v1/error-events", strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", testCase.contentType)
			response := httptest.NewRecorder()
			handler.HandlePostErrorEvent(response, request)
			if response.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, testCase.wantStatus)
			}
			if !strings.Contains(response.Body.String(), testCase.wantBody) {
				t.Fatalf("body %q does not contain %q", response.Body.String(), testCase.wantBody)
			}
			if strings.Contains(response.Body.String(), "secret database detail") {
				t.Fatal("response leaked internal error details")
			}
		})
	}
}
