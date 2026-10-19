package realtime

import "testing"

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
