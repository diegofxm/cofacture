// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

// samplePOSCreditAdjustment builds a "Nota de Ajuste de tipo crédito al Documento Equivalente"
// (CreditNoteTypeCode/DocumentTypeCode "94", Technical Annex Documento Equivalente Electrónico
// V1.0 sections 16.3/16.4.2) referencing the POS ticket from samplePOSInvoice(). Reuses
// sampleCreditNote()'s shape and overrides only what the annex actually changes: ProfileID,
// DocumentTypeCode/CreditNoteTypeCode, and OperationTypeCode/CustomizationID -- which for a
// Documento Equivalente adjustment note is the REFERENCED document's own InvoiceTypeCode ("20"
// for POS, section 16.4.2), not a generic "references some invoice" code like a normal Credit
// Note uses. Do not confuse this "20" with the unrelated "20" OperationTypeCode a regular
// Credit Note uses in sampleCreditNote() -- same digits, two different DIAN catalogs.
func samplePOSCreditAdjustment() domain.CreditNote {
	cn := sampleCreditNote()
	cn.Invoice.ProfileID = "DIAN 2.1: Nota de ajuste crédito al documento equivalente"
	cn.Invoice.OperationTypeCode = "20" // CustomizationID: referenced document is POS ("20")
	cn.Invoice.DocumentTypeCode = "94"
	cn.CreditNoteTypeCode = "94"
	cn.BillingReference = domain.BillingReference{
		Prefix:    "SETP",
		Number:    "1",
		CUFE:      "db07502cd11c006f4666e2e299fd77e5a47bd790d9da18786dace4b4d0c4b8972643843e8b7444fe23a0fc8aa1fdf5f2",
		IssueDate: "2024-01-20",
		HashType:  "CUDE-SHA384", // references a Documento Equivalente Electrónico (POS), not an Invoice
	}
	return cn
}

// samplePOSDebitAdjustment is the "Nota de Ajuste de tipo débito" (DocumentTypeCode "93")
// sibling of samplePOSCreditAdjustment — see that function's doc comment.
func samplePOSDebitAdjustment() domain.DebitNote {
	dn := sampleDebitNote()
	dn.Invoice.ProfileID = "DIAN 2.1: Nota de ajuste débito al documento equivalente"
	dn.Invoice.OperationTypeCode = "20" // CustomizationID: referenced document is POS ("20")
	dn.Invoice.DocumentTypeCode = "93"
	dn.BillingReference = domain.BillingReference{
		Prefix:    "SETP",
		Number:    "1",
		CUFE:      "db07502cd11c006f4666e2e299fd77e5a47bd790d9da18786dace4b4d0c4b8972643843e8b7444fe23a0fc8aa1fdf5f2",
		IssueDate: "2024-01-20",
		HashType:  "CUDE-SHA384", // references a Documento Equivalente Electrónico (POS), not an Invoice
	}
	return dn
}

// Both tests below are structural checks only, like every other golden test in this package —
// not yet confirmed against DIAN's certification environment. See pos_test.go's doc comment for
// the same caveat.

func TestBuildCreditNote_POSAdjustment_Golden(t *testing.T) {
	doc, err := BuildCreditNote(samplePOSCreditAdjustment())
	if err != nil {
		t.Fatalf("BuildCreditNote: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/pos_credit_adjustment_golden.xml"

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

func TestBuildDebitNote_POSAdjustment_Golden(t *testing.T) {
	doc, err := BuildDebitNote(samplePOSDebitAdjustment())
	if err != nil {
		t.Fatalf("BuildDebitNote: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/pos_debit_adjustment_golden.xml"

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
