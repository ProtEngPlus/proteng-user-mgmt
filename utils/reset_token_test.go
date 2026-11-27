package utils

import "testing"

func TestHashResetToken(t *testing.T) {
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

	if got := HashResetToken("abc"); got != want {
		t.Errorf("HashResetToken() = %s, want %s", got, want)
	}
	if got := HashResetToken("abc"); got == Encode("abc") {
		t.Error("hash must not be the reversible encoding")
	}
}
