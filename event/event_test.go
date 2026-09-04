// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package event

import "testing"

// TestCompute reproduces the official worked example from Technical Annex 1.9, section 11.5.1.
func TestCompute(t *testing.T) {
	const want = "0d91ba25b01f5e7dbda870a11b274501d3a62a73e91932c473c86c93f12a142a2ac45876efcde3e679024a01c0be41f9"

	got := Compute(
		"1",              // Num_DE
		"2019-04-30",     // Fec_Emi
		"19:48:50-05:00", // Hor_Emi
		"99998888",       // NitFE
		"800197268",      // DocAdq
		"030",            // ResponseCode
		"FE123",          // ID
		"01",             // DocumentTypeCode
		"11111",          // Software-PIN
	)

	if got != want {
		t.Errorf("Compute() = %s, want %s", got, want)
	}
}
