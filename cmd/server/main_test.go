package main

import "testing"

func TestRandomNonce(t *testing.T) {
	n1, err := randomNonce(16)
	if err != nil {
		t.Fatalf("nonce generation failed: %v", err)
	}
	n2, err := randomNonce(16)
	if err != nil {
		t.Fatalf("nonce generation failed: %v", err)
	}
	if len(n1) != 32 || len(n2) != 32 {
		t.Fatalf("unexpected nonce length: %d, %d", len(n1), len(n2))
	}
	if n1 == n2 {
		t.Fatalf("expected unique nonces")
	}
}
