// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

// sampleAdjustmentNote builds an Adjustment Note to the Support Document (InvoiceTypeCode
// "95"), reusing sampleSupportDocument()'s roles (Supplier = SNO, Customer = issuer) plus a
// BillingReference to the original Support Document. ProfileID/OperationTypeCode/Prefix below
// are not a guess — they are the exact field values used in a real Adjustment Note that DIAN's
// habilitación environment accepted (StatusCode 00, "ha sido autorizada"); see
// integration-tests/adjustment_note_test.go.
func sampleAdjustmentNote() domain.AdjustmentNote {
	inv := sampleSupportDocument()
	inv.ProfileID = "DIAN 2.1: Nota de ajuste al documento soporte en adquisiciones efectuadas a sujetos no obligados a expedir factura o documento equivalente"
	inv.DocumentTypeCode = "95"
	inv.HashType = "CUDS-SHA384"
	inv.Prefix = "NAP"
	inv.Number = "1"

	return domain.AdjustmentNote{
		Invoice: inv,
		BillingReference: domain.BillingReference{
			Prefix:    "SEDS",
			Number:    "1",
			CUFE:      "18015e1f4f6b1eb55cf6d5eaa1f752bed3b0402e0cf11eb515c1ce5ccbe9bca120cd4776ee3b1e5c281e0fd2711d40d1",
			IssueDate: "2024-01-20",
			HashType:  "CUDS-SHA384", // references the original Support Document (unused by appendNABillingReference, which hardcodes this correctly — set here for clarity)
		},
		DiscrepancyResponse: &domain.DiscrepancyResponse{
			ReferenceID:  "SEDS1",
			ResponseCode: "1",
			Description:  "Corrección de valores",
		},
	}
}

// TestBuildAdjustmentNote_Golden is the first dedicated test for this builder function — until
// now it was only exercised by integration-tests/adjustment_note_test.go (real DIAN submission,
// gated behind COFACTURE_TEST_FIXTURES_DIR), never by go test ./... inside this repo.
func TestBuildAdjustmentNote_Golden(t *testing.T) {
	doc, err := BuildAdjustmentNote(sampleAdjustmentNote())
	if err != nil {
		t.Fatalf("BuildAdjustmentNote: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/adjustment_note_golden.xml"

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
