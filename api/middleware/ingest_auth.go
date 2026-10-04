// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"
)

// IngestAuth rejects requests without the configured server-side bearer credential.
func IngestAuth(expectedToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		scheme, providedToken, ok := strings.Cut(authorization, " ")
		if expectedToken == "" || !ok || !strings.EqualFold(scheme, "Bearer") || providedToken == "" {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		expectedHash := sha256.Sum256([]byte(expectedToken))
		providedHash := sha256.Sum256([]byte(providedToken))
		if subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) != 1 {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
