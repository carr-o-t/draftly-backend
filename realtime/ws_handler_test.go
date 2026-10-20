package realtime

import (
	"net/http/httptest"
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	t.Setenv("FRONTEND_ORIGIN", "https://app.example.com")

	cases := []struct {
		name   string
		origin string
		want   bool
	}{
		{"allowed origin", "https://app.example.com", true},
		{"other origin", "https://evil.example.com", false},
		{"no origin header", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/ws/x", nil)
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			if got := checkOrigin(r); got != tc.want {
				t.Fatalf("origin %q: got %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	cases := []struct {
		name   string
		header string
		want   string
	}{
		{"standard", "Bearer abc123", "abc123"},
		{"lowercase scheme", "bearer abc123", "abc123"},
		{"extra spaces", "Bearer   abc123  ", "abc123"},
		{"wrong scheme", "Basic abc123", ""},
		{"missing scheme", "abc123", ""},
		{"empty", "", ""},
		{"scheme only", "Bearer", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := bearerToken(tc.header); got != tc.want {
				t.Fatalf("bearerToken(%q) = %q, want %q", tc.header, got, tc.want)
			}
		})
	}
}
