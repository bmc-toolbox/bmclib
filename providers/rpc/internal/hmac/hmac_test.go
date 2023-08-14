package hmac

import "testing"

func TestNewHMAC(t *testing.T) {
	h := NewHMAC()
	if h.Hashes == nil {
		t.Fatal("expected Hashes to be initialized")
	}
	if !h.PrefixSig {
		t.Fatal("expected NoPrefix to be false")
	}
}
