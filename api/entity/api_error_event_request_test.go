package entity

import "testing"

func TestPostErrorEventRequestValidateAndParse(t *testing.T) {
	testCaseList := []struct {
		name        string
		requestData PostErrorEventRequest
		wantError   bool
	}{
		{
			name: "valid event",
			requestData: PostErrorEventRequest{
				Service: "buildlog-frontend", Environment: "production", Status: 500,
				Method: "get", URL: "https://example.com/contact", Path: "/contact",
				CommitSHA: "0123456789abcdef0123456789abcdef01234567",
			},
		},
		{
			name: "rejects query string",
			requestData: PostErrorEventRequest{
				Service: "buildlog-frontend", Environment: "production", Status: 500,
				Method: "GET", URL: "https://example.com/contact?token=value", Path: "/contact",
			},
			wantError: true,
		},
		{
			name: "rejects non-server error",
			requestData: PostErrorEventRequest{
				Service: "buildlog-frontend", Environment: "production", Status: 404,
				Method: "GET", URL: "https://example.com/contact", Path: "/contact",
			},
			wantError: true,
		},
		{
			name: "rejects path mismatch",
			requestData: PostErrorEventRequest{
				Service: "buildlog-frontend", Environment: "production", Status: 500,
				Method: "GET", URL: "https://example.com/contact", Path: "/other",
			},
			wantError: true,
		},
		{
			name: "rejects non-https URL",
			requestData: PostErrorEventRequest{
				Service: "buildlog-frontend", Environment: "production", Status: 500,
				Method: "GET", URL: "http://example.com/contact", Path: "/contact",
			},
			wantError: true,
		},
	}

	for _, testCase := range testCaseList {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.requestData.ValidateAndParse()
			if (err != nil) != testCase.wantError {
				t.Fatalf("ValidateAndParse() error = %v, wantError %v", err, testCase.wantError)
			}
			if err == nil && testCase.requestData.Method != "GET" {
				t.Errorf("method = %q, want GET", testCase.requestData.Method)
			}
		})
	}
}
