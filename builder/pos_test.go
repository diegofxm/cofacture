// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

// samplePOSInvoice reuses sampleInvoice() and overrides only the four fields the Documento
// Equivalente Electrónico Technical Annex V1.0 (section 8.2/16.3/16.4.1) actually changes for
// the "tiquete de máquina registradora con sistema P.O.S." document (InvoiceTypeCode 20):
// ProfileID, CustomizationID (OperationTypeCode), DocumentTypeCode, and HashType. Everything
// else — Supplier/Customer/Lines/Totals/NumberingRange/SoftwareProvider shape — is identical to
// a regular invoice per the annex's own field tables (see cude.Compute's doc comment for the
// CUDE side of this same claim). This test exists to verify that claim holds for BuildInvoice
// too, not just assert it in a comment.
func samplePOSInvoice() domain.Invoice {
	inv := sampleInvoice()
	inv.ProfileID = "DIAN 2.1: Documento Equivalente POS"
	inv.OperationTypeCode = "10"
	inv.DocumentTypeCode = "20"
	inv.HashType = "CUDE-SHA384"
	return inv
}

// TestBuildInvoice_POS_Golden is a structural check only (like every other golden test in this
// package): it has not been confirmed against DIAN's certification environment the way the
// regular Invoice/CreditNote/DebitNote/SupportDocument builders were. Treat it as "should be
// correct per the annex" until a real Documento Equivalente Electrónico has been submitted and
// accepted — same caveat as builder/event.go for RADIAN.
func TestBuildInvoice_POS_Golden(t *testing.T) {
	doc, err := BuildInvoice(samplePOSInvoice())
	if err != nil {
		t.Fatalf("BuildInvoice: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/pos_invoice_golden.xml"

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
