// SPDX-License-Identifier: Apache-2.0
package library

import "testing"

func TestValidateDatabaseTransport(t *testing.T) {
	testCaseList := []struct {
		name        string
		databaseURL string
		environment string
		wantError   bool
	}{
		{name: "allows local development URL", databaseURL: "postgres://localhost/db?sslmode=disable", environment: "development"},
		{name: "requires TLS in production", databaseURL: "postgres://db.example/db?sslmode=require", environment: "production"},
		{name: "rejects plaintext production URL", databaseURL: "postgres://db.example/db?sslmode=disable", environment: "production", wantError: true},
		{name: "rejects missing production ssl mode", databaseURL: "postgres://db.example/db", environment: "production", wantError: true},
		{name: "rejects unknown environment", databaseURL: "postgres://localhost/db", environment: "prod", wantError: true},
	}

	for _, testCase := range testCaseList {
		t.Run(testCase.name, func(t *testing.T) {
			err := ValidateDatabaseTransport(testCase.databaseURL, testCase.environment)
			if (err != nil) != testCase.wantError {
				t.Fatalf("ValidateDatabaseTransport() error = %v, wantError %v", err, testCase.wantError)
			}
		})
	}
}

func TestValidateIngestToken(t *testing.T) {
	if err := ValidateIngestToken("0123456789abcdef0123456789abcdef"); err != nil {
		t.Fatalf("ValidateIngestToken() error = %v", err)
	}
	if err := ValidateIngestToken("short"); err == nil {
		t.Fatal("ValidateIngestToken() error = nil for a short token")
	}
}
