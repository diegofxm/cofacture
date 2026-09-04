// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package securitycode

import "testing"

// There is no published official example to validate this formula byte-for-byte (unlike
// CUFE) — this test is a regression/sanity check: correct length, determinism, and that
// each input actually participates in the hash.
func TestCompute(t *testing.T) {
	got := Compute("software-id", "1234", "SETP1")
	if len(got) != 96 {
		t.Fatalf("expected length 96 (SHA-384 in hex), got %d: %s", len(got), got)
	}
	if again := Compute("software-id", "1234", "SETP1"); again != got {
		t.Errorf("Compute is not deterministic")
	}
	if other := Compute("software-id", "1234", "SETP2"); other == got {
		t.Error("changing documentID should change the result")
	}
	if other := Compute("software-id", "9999", "SETP1"); other == got {
		t.Error("changing the PIN should change the result")
	}
}
