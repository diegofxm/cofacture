// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package payroll

import "testing"

// TestXMLFileName, TestAdjustXMLFileName and TestZIPFileName lock in the exact file-naming
// convention from section 3.3/3.5 of the Electronic Payroll Technical Annex: prefix, NIT
// left-padded to 10 digits, year as 2 digits, consecutive as 8 uppercase hex digits.
func TestXMLFileName(t *testing.T) {
	got := XMLFileName("6382356", 2024, 1)
	want := "nie000638235624" + "00000001" + ".xml"
	if got != want {
		t.Errorf("XMLFileName() = %q, want %q", got, want)
	}
}

func TestAdjustXMLFileName(t *testing.T) {
	got := AdjustXMLFileName("6382356", 2024, 1)
	want := "niae000638235624" + "00000001" + ".xml"
	if got != want {
		t.Errorf("AdjustXMLFileName() = %q, want %q", got, want)
	}
}

func TestZIPFileName(t *testing.T) {
	got := ZIPFileName("6382356", 2024, 1)
	want := "z000638235624" + "00000001" + ".zip"
	if got != want {
		t.Errorf("ZIPFileName() = %q, want %q", got, want)
	}
}
