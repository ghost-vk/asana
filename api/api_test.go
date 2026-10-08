package api

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGetURL(t *testing.T) {
	if getURL("/", url.Values{}) != "https://app.asana.com/" {
		t.Error("built URL is Invalid")
	}

	params := url.Values{}
	params.Add("hoge", "1")
	if getURL("/wei", params) != "https://app.asana.com/wei?hoge=1" {
		t.Error("built URL is Invalid")
	}
}

func TestSendRetriesOnRateLimit(t *testing.T) {
	originalTransport, originalWait := http.DefaultTransport, retryAfter
	defer func() { http.DefaultTransport, retryAfter = originalTransport, originalWait }()
	retryAfter = func(string) time.Duration { return 0 }

	calls := 0
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		body, _ := io.ReadAll(req.Body)
		if string(body) != `{"data":{}}` {
			t.Fatalf("attempt %d body = %q, the retry must resend it", calls, body)
		}
		status := http.StatusOK
		if calls < 3 {
			status = http.StatusTooManyRequests
		}
		return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
	})
	if _, err := Request("PUT", "/tasks/1", `{"data":{}}`); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}
