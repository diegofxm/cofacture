// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"os"
	"testing"
)

// TestBuildInvoice_POS_Rounding_Golden confirms PayableRoundingAmount is serialized (with its
// negative sign preserved) when Totals.RoundingCents is non-zero, and that it appears between
// PrepaidAmount and PayableAmount — same element order the annex documents.
func TestBuildInvoice_POS_Rounding_Golden(t *testing.T) {
	inv := samplePOSInvoice()
	inv.Totals.RoundingCents = -50_00 // -50.00 pesos, a downward rounding adjustment

	doc, err := BuildInvoice(inv)
	if err != nil {
		t.Fatalf("BuildInvoice: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/pos_invoice_rounding_golden.xml"

	if *update {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Skip("golden file regenerated, review by hand before trusting it")
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (missing -update run?): %v", err)
	}
	if got != string(want) {
		t.Errorf("generated XML does not match %s\n--- got ---\n%s", goldenPath, got)
	}
}

// TestBuildInvoice_NoRounding_OmitsPayableRoundingAmount confirms the element is entirely absent
// when RoundingCents is left at its zero value — same optional-emission convention as
// PrepaidAmount.
func TestBuildInvoice_NoRounding_OmitsPayableRoundingAmount(t *testing.T) {
	doc, err := BuildInvoice(sampleInvoice())
	if err != nil {
		t.Fatalf("BuildInvoice: %v", err)
	}
	if el := doc.FindElement("//cbc:PayableRoundingAmount"); el != nil {
		t.Errorf("PayableRoundingAmount should be absent when RoundingCents == 0, got %q", el.Text())
	}
}
