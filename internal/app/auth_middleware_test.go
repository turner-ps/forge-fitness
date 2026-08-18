package app

import "testing"

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantToken  string
		wantParsed bool
	}{
		{name: "valid", header: "Bearer firebase-token", wantToken: "firebase-token", wantParsed: true},
		{name: "case insensitive scheme", header: "bearer firebase-token", wantToken: "firebase-token", wantParsed: true},
		{name: "missing header", header: "", wantParsed: false},
		{name: "wrong scheme", header: "Basic firebase-token", wantParsed: false},
		{name: "missing token", header: "Bearer", wantParsed: false},
		{name: "token contains spaces", header: "Bearer firebase token", wantParsed: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, parsed := bearerToken(test.header)
			if token != test.wantToken || parsed != test.wantParsed {
				t.Fatalf("bearerToken(%q) = (%q, %t), want (%q, %t)", test.header, token, parsed, test.wantToken, test.wantParsed)
			}
		})
	}
}
