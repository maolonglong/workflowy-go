package workflowy

import "testing"

func TestServiceNewRequestPreservesQueryString(t *testing.T) {
	s := &service{client: &Client{
		baseURL:   defaultBaseURL,
		apiKey:    "test-api-key",
		userAgent: defaultUserAgent,
	}}

	req, err := s.newRequest("GET", "/nodes?parent_id=inbox", nil)
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}

	if got, want := req.URL.Path, "/api/v1/nodes"; got != want {
		t.Fatalf("request path = %q, want %q", got, want)
	}

	if got, want := req.URL.Query().Get("parent_id"), "inbox"; got != want {
		t.Fatalf("parent_id query = %q, want %q", got, want)
	}
}
