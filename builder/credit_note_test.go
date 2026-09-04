// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

func sampleCreditNote() domain.CreditNote {
	inv := sampleInvoice()
	inv.OperationTypeCode = "20" // credit note referencing a specific invoice
	inv.DocumentTypeCode = "91"
	inv.HashType = "CUDE-SHA384"
	inv.Prefix = "SETPNC"
	inv.Number = "1"

	return domain.CreditNote{
		Invoice:            inv,
		CreditNoteTypeCode: "91", // fixed DIAN code for a Credit Note (document type); the List 22 concept goes in DiscrepancyResponse
		BillingReference: domain.BillingReference{
			Prefix:    "SETP",
			Number:    "1",
			CUFE:      "8bb918b19ba22a694f1da11c643b5e9de39adf60311cf179179e9b33381030bcd4c3c3f156c506ed5908f9276f5bd9b4",
			IssueDate: "2024-01-20",
			HashType:  "CUFE-SHA384", // references a regular Invoice
		},
		DiscrepancyResponse: &domain.DiscrepancyResponse{
			ReferenceID:  "SETP1",
			ResponseCode: "2",
			Description:  "Anulación de factura electrónica",
		},
	}
}

func TestBuildCreditNote_Golden(t *testing.T) {
	doc, err := BuildCreditNote(sampleCreditNote())
	if err != nil {
		t.Fatalf("BuildCreditNote: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/credit_note_golden.xml"

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
