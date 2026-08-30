package controller

import (
	"net/http/httptest"
	"testing"
)

func TestIsSameOriginOrNonBrowser(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "non-browser client", want: true},
		{name: "same origin", origin: "https://v2raya.example:2017", want: true},
		{name: "different origin", origin: "https://evil.example", want: false},
		{name: "invalid origin", origin: "://bad", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://v2raya.example:2017/api/live-flow", nil)
			req.Host = "v2raya.example:2017"
			if test.origin != "" {
				req.Header.Set("Origin", test.origin)
			}
			if got := isSameOriginOrNonBrowser(req); got != test.want {
				t.Errorf("isSameOriginOrNonBrowser() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := &RateLimiter{clients: make(map[string]*ClientInfo)}
	for attempt := 0; attempt < rateLimitMaximum; attempt++ {
		if !limiter.IsAllowed("client") {
			t.Fatalf("attempt %d should be allowed", attempt+1)
		}
	}
	if limiter.IsAllowed("client") {
		t.Fatal("attempt above the limit should be rejected")
	}
}
