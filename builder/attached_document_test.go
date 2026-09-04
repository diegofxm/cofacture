// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

func sampleAttachedDocument() domain.AttachedDocument {
	return domain.AttachedDocument{
		EnvironmentCode:  "2",
		ID:               "1",
		IssueDate:        "2024-01-20",
		IssueTime:        "10:05:00-05:00",
		ParentDocumentID: "SETP1",

		Sender: domain.AttachedPartyInfo{
			Name:           "MI EMPRESA S.A.S.",
			Identification: domain.Identification{Number: "900123456", TypeCode: "31", VerificationCode: "3"},
			TaxRegimeCode:  "48",
			LiabilityCodes: []string{"O-13", "O-15", "O-23"},
			TaxSchemeCode:  "01",
			TaxSchemeName:  "IVA",
		},
		Receiver: domain.AttachedPartyInfo{
			Name:           "Juan Pérez",
			Identification: domain.Identification{Number: "1234567890", TypeCode: "13"},
			TaxSchemeCode:  "01",
			TaxSchemeName:  "IVA",
		},

		AttachmentXML: "<Invoice>contenido simplificado para esta prueba, no se valida aquí</Invoice>",

		ValidationResults: []domain.ValidationResult{
			{
				LineID:                 "1",
				DocumentID:             "SETP1",
				DocumentCUFE:           "8bb918b19ba22a694f1da11c643b5e9de39adf60311cf179179e9b33381030bcd4c3c3f156c506ed5908f9276f5bd9b4",
				DocumentHashType:       "CUFE-SHA384",
				DocumentIssueDate:      "2024-01-20",
				ApplicationResponseXML: "<ApplicationResponse>contenido simplificado para esta prueba</ApplicationResponse>",
				ValidatorID:            "Unidad Especial Dirección de Impuestos y Aduanas Nacionales",
				ValidationResultCode:   "02",
				ValidationDate:         "2024-01-20",
				ValidationTime:         "10:10:00-05:00",
			},
		},
	}
}

func TestBuildInvoiceAttachedDocument_Golden(t *testing.T) {
	doc, err := BuildInvoiceAttachedDocument(sampleAttachedDocument())
	if err != nil {
		t.Fatalf("BuildInvoiceAttachedDocument: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/attached_document_golden.xml"

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
