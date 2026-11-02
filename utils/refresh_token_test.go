package utils

import "testing"

func TestNewRefreshToken(t *testing.T) {
	raw, hash, err := NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken() error = %v", err)
	}
	if len(raw) != 43 {
		t.Errorf("raw length = %d, want 43", len(raw))
	}
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}
	if hash != HashRefreshToken(raw) {
		t.Errorf("hash does not match HashRefreshToken(raw)")
	}

	other, _, err := NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken() error = %v", err)
	}
	if other == raw {
		t.Errorf("two calls returned the same token")
	}
}

func TestHashRefreshToken(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"known sha256", "abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"empty", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HashRefreshToken(tc.input); got != tc.want {
				t.Errorf("HashRefreshToken(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
