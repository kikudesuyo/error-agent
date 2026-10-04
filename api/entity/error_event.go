// SPDX-License-Identifier: Apache-2.0
package entity

// ErrorEvent contains validated and redacted data safe to persist for investigation.
type ErrorEvent struct {
	Service     string
	Environment string
	Status      int
	Method      string
	URL         string
	Path        string
	ErrorType   string
	Message     string
	Stack       string
	CommitSHA   string
	CFRay       string
	UserAgent   string
}
