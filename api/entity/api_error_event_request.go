// SPDX-License-Identifier: Apache-2.0
package entity

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	serviceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)
	commitSHAPattern   = regexp.MustCompile(`(?i)^[a-f0-9]{7,64}$`)
)

// PostErrorEventRequest is the lower-camel-case event contract used by Buildlog.
type PostErrorEventRequest struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Status      int    `json:"status"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	Path        string `json:"path"`
	ErrorType   string `json:"errorType"`
	Message     string `json:"message"`
	Stack       string `json:"stack"`
	CommitSHA   string `json:"commitSha"`
	CFRay       string `json:"cfRay"`
	UserAgent   string `json:"userAgent"`
}

// ValidateAndParse rejects event data outside the documented contract.
func (request *PostErrorEventRequest) ValidateAndParse() error {
	request.Service = strings.TrimSpace(request.Service)
	request.Environment = strings.TrimSpace(request.Environment)
	request.Method = strings.ToUpper(strings.TrimSpace(request.Method))
	request.Path = strings.TrimSpace(request.Path)
	request.URL = strings.TrimSpace(request.URL)
	request.ErrorType = strings.TrimSpace(request.ErrorType)

	if !serviceNamePattern.MatchString(request.Service) {
		return fmt.Errorf("service is invalid")
	}
	if request.Environment == "" || len(request.Environment) > 64 {
		return fmt.Errorf("environment is invalid")
	}
	if request.Status < 500 || request.Status > 599 {
		return fmt.Errorf("status must be between 500 and 599")
	}
	if !isAllowedMethod(request.Method) {
		return fmt.Errorf("method is invalid")
	}
	if len(request.Path) == 0 || len(request.Path) > 2048 || !strings.HasPrefix(request.Path, "/") {
		return fmt.Errorf("path is invalid")
	}
	parsedURL, err := url.ParseRequestURI(request.URL)
	if err != nil || parsedURL.Host == "" || parsedURL.Scheme != "https" {
		return fmt.Errorf("url is invalid")
	}
	if parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return fmt.Errorf("url must not contain credentials, query, or fragment")
	}
	if len(request.ErrorType) > 128 || len(request.Message) > 8192 || len(request.Stack) > 32768 {
		return fmt.Errorf("error details exceed the allowed size")
	}
	if len(request.CommitSHA) > 64 || len(request.CFRay) > 128 || len(request.UserAgent) > 1024 {
		return fmt.Errorf("metadata exceeds the allowed size")
	}
	if request.CommitSHA != "" && !commitSHAPattern.MatchString(request.CommitSHA) {
		return fmt.Errorf("commitSha is invalid")
	}
	if parsedURL.EscapedPath() != "" && parsedURL.EscapedPath() != request.Path {
		return fmt.Errorf("url path does not match path")
	}
	return nil
}

func isAllowedMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS":
		return true
	default:
		return false
	}
}
