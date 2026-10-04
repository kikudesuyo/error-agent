// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIngestAuth(t *testing.T) {
	testCaseList := []struct {
		name          string
		token         string
		authorization string
		wantStatus    int
	}{
		{name: "accepts valid token", token: "expected-token", authorization: "Bearer expected-token", wantStatus: http.StatusNoContent},
		{name: "rejects invalid token", token: "expected-token", authorization: "Bearer wrong-token", wantStatus: http.StatusUnauthorized},
		{name: "rejects missing token", token: "expected-token", wantStatus: http.StatusUnauthorized},
		{name: "rejects unset credential", authorization: "Bearer expected-token", wantStatus: http.StatusUnauthorized},
	}

	for _, testCase := range testCaseList {
		t.Run(testCase.name, func(t *testing.T) {
			next := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodPost, "/v1/error-events", nil)
			request.Header.Set("Authorization", testCase.authorization)
			response := httptest.NewRecorder()
			IngestAuth(testCase.token, next).ServeHTTP(response, request)
			if response.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, testCase.wantStatus)
			}
		})
	}
}
