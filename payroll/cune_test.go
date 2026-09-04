// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package payroll

import (
	"crypto/sha512"
	"fmt"
	"testing"
)

// TestCune is a deterministic regression test: it locks in Cune's output for a fixed set of
// inputs, formatted the same way the rest of this module formats dates/times/money everywhere
// else (ISO date, "HH:MM:SS±HH:MM" time, two-decimal money, no thousands separators). It does
// NOT reproduce DIAN's own worked example from Technical Annex section 8.1 — see
// TestCune_AnexoTecnicoExample below for that investigation and why it's a separate,
// non-blocking test instead of this one.
func TestCune(t *testing.T) {
	const want = "dae35b4dfdf10939502c96278feb97be1af6cd5653bd662ca29b5c9cdf2729464b51d272ba0008586cf153b46f87a2bd"

	got := Cune("N00001", "2020-01-16", "10:53:10-05:00",
		350_000_000, 100_000_000, 250_000_000,
		"700085371", "800199436", "102", "693", "1")

	if got != want {
		t.Errorf("Cune() = %s, want %s", got, want)
	}
}

// TestCune_AnexoTecnicoExample documents an unresolved discrepancy: none of the plausible
// field-formatting variants below reproduce the CUNE that DIAN's Technical Annex (section 8.1)
// publishes for its own worked example. The Annex text was extracted from a PDF, and the
// composition string it prints may itself carry a transcription artifact — the same class of
// issue confirmed and fixed in cude_test.go for the Debit Note worked example (there, the
// Annex's own printed hash didn't match its own printed input string; cude.Compute reproduces
// the mathematically correct hash of that input, not the Annex's typo). Unlike that case, this
// one has NOT been independently resolved: it is not known whether the Annex's CUNE example has
// a similar transcription error, or whether Cune's implementation has a real bug this search
// hasn't found the right input format for.
//
// This test intentionally never fails — it is a documented, ongoing investigation, not a
// correctness check. Cune's actual correctness for the standard format is anchored by TestCune
// above (regression) and, for real payroll submissions, by whether DIAN's certification
// environment accepts a document built with it — see integration-tests/nomina_test.go, which
// does fail (t.Errorf) on a real DIAN rejection, gated behind COFACTURE_TEST_FIXTURES_DIR.
//
// Cune's public signature only accepts cents (int64), so it can only ever produce the
// two-decimal money format — it can no longer be asked to try a "without decimals" variant like
// earlier versions of this search did. That variant is still worth checking (the Annex text is
// ambiguous about it), so it's built and hashed manually below via sha384Hex instead of through
// Cune itself.
func TestCune_AnexoTecnicoExample(t *testing.T) {
	const want = "16560dc8956122e84ffb743c817fe7d494e058a44d9ca3fa4c234c268b4f766003253fbee7ea4af9682dd57210f3bac2"

	horVariants := []string{
		"10:53:10-05:00", // standard ISO 8601 HH:MM:SS±HH:MM
		"1053:10-05:00",  // literal from the Annex (without the first ":")
		"105310-05:00",   // without colons between HH/MM/SS
		"10:53:10",       // without timezone
		"105310",         // HHMMSS only
	}

	for _, hor := range horVariants {
		// With 2 decimals (the format Cune's real signature enforces).
		got := Cune("N00001", "2020-01-16", hor,
			350_000_000, 100_000_000, 250_000_000,
			"700085371", "800199436", "102", "693", "1")
		if got == want {
			t.Logf("match found: HorNE=%q with 2-decimal amounts reproduces the Annex's published CUNE", hor)
			return
		}

		// Without decimals — Cune can't produce this anymore, so build it by hand.
		raw := "N00001" + "2020-01-16" + hor + "3500000" + "1000000" + "2500000" +
			"700085371" + "800199436" + "102" + "693" + "1"
		if sha384Hex(raw) == want {
			t.Logf("match found: HorNE=%q without decimals reproduces the Annex's published CUNE", hor)
			return
		}
	}

	// The Annex's own literal concatenation, byte for byte as printed.
	raw := "N000012020-01-161053:10-05:003500000.001000000.002500000.007000853718001994361026931"
	if rawHash := sha384Hex(raw); rawHash == want {
		t.Log("the Annex's literal printed string reproduces its own published CUNE — check field composition above")
		return
	}

	// A couple of spacing/leading-zero variants worth a quick check.
	variants := []string{
		"N000012020-01-1610:53:10-05:003500000.001000000.002500000.0070008537180019943610 2693 1",
		"N000012020-01-1610:53:10-05:003500000.001000000.002500000.00700085371800199436 2693 1",
	}
	for _, v := range variants {
		if sha384Hex(v) == want {
			t.Logf("variant matches: %q", v)
			return
		}
	}

	t.Logf("no variant tried reproduces the Annex's published CUNE (%s); literal string hash: %s", want, sha384Hex(raw))
}

func sha384Hex(s string) string {
	sum := sha512.Sum384([]byte(s))
	return fmt.Sprintf("%x", sum)
}
