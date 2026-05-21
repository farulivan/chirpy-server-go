package auth

import (
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "Valid Bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer valid_token"},
			},
			wantToken: "valid_token",
			wantErr:   false,
		},
		{
			name:      "Missing Authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "Malformed Authorization header",
			headers: http.Header{
				"Authorization": []string{"InvalidBearer token"},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
	t.Run("Returns no error", func(t *testing.T) {
		if _, err := MakeRefreshToken(); err != nil {
			t.Errorf("MakeRefreshToken() error = %v", err)
		}
	})

	t.Run("Returns 64 hex characters (32 random bytes)", func(t *testing.T) {
		got, err := MakeRefreshToken()
		if err != nil {
			t.Fatalf("MakeRefreshToken() error = %v", err)
		}
		if len(got) != 64 {
			t.Errorf("MakeRefreshToken() len = %d, want 64", len(got))
		}
	})

	t.Run("Returns valid hex", func(t *testing.T) {
		got, err := MakeRefreshToken()
		if err != nil {
			t.Fatalf("MakeRefreshToken() error = %v", err)
		}
		if _, err := hex.DecodeString(got); err != nil {
			t.Errorf("MakeRefreshToken() returned invalid hex %q: %v", got, err)
		}
	})

	t.Run("Returns lowercase hex only", func(t *testing.T) {
		got, err := MakeRefreshToken()
		if err != nil {
			t.Fatalf("MakeRefreshToken() error = %v", err)
		}
		if got != strings.ToLower(got) {
			t.Errorf("MakeRefreshToken() = %q, want lowercase hex", got)
		}
	})

	t.Run("Produces unique tokens across calls", func(t *testing.T) {
		const n = 100
		seen := make(map[string]struct{}, n)
		for i := 0; i < n; i++ {
			tok, err := MakeRefreshToken()
			if err != nil {
				t.Fatalf("MakeRefreshToken() error = %v", err)
			}
			if _, dup := seen[tok]; dup {
				t.Fatalf("MakeRefreshToken() produced duplicate token %q after %d calls", tok, i)
			}
			seen[tok] = struct{}{}
		}
	})
}
