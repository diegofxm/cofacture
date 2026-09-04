// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

func sampleSupportDocument() domain.Invoice {
	sup := sampleInvoice()
	sup.OperationTypeCode = "10" // Resident
	sup.DocumentTypeCode = "05"
	sup.HashType = "CUDS-SHA384"
	sup.ProfileID = "DIAN 2.1: documento soporte en adquisiciones efectuadas a no obligados a facturar."
	sup.Prefix = "DS"
	sup.Number = "1"

	// Roles reversed: Supplier = third party not obligated to invoice (SNO), Customer = issuer.
	// DIAN requires schemeName="31" (NIT) for the SNO — verified in DS-real.xml.
	sup.Supplier = domain.Party{
		EntityTypeCode: "2",
		Identification: domain.Identification{
			Number:           "1020304050",
			TypeCode:         "31",
			VerificationCode: "8",
		},
		Name:           "María García",
		TaxSchemeCode:  "ZZ",
		TaxSchemeName:  "No aplica",
		LiabilityCodes: []string{"R-99-PN"},
		TaxRegimeCode:  "49",
		Address: domain.Address{
			Line:        "Vereda El Rosal",
			CityCode:    "05001",
			CityName:    "Medellín",
			StateCode:   "05",
			StateName:   "Antioquia",
			CountryCode: "CO",
			CountryName: "Colombia",
		},
	}
	// Customer = the issuing company (same data as the original Supplier in sampleInvoice).
	sup.Customer = domain.Party{
		EntityTypeCode: "2",
		Identification: domain.Identification{
			Number:           "900123456",
			TypeCode:         "31",
			VerificationCode: "8",
		},
		Name: "MI EMPRESA S.A.S.",
		Address: domain.Address{
			Line:        "Calle 123 #45-67",
			CityCode:    "11001",
			CityName:    "Bogotá D.C.",
			StateCode:   "11",
			StateName:   "Bogotá D.C.",
			CountryCode: "CO",
			CountryName: "Colombia",
		},
		TaxRegimeCode:  "48",
		LiabilityCodes: []string{"O-13", "O-15", "O-23"},
		TaxSchemeCode:  "01",
		TaxSchemeName:  "IVA",
	}

	// A ReteRenta withholding of 3.5% on a base of 126,050,420.
	sup.WithholdingTaxes = []domain.Tax{
		{
			TypeCode:           "06",
			TypeName:           "ReteRenta",
			TaxableAmountCents: 126_050_420,
			TaxAmountCents:     4_411_765,
			Percent:            3.5,
		},
	}

	return sup
}

func TestBuildSupportDocument_Golden(t *testing.T) {
	doc, err := BuildSupportDocument(sampleSupportDocument())
	if err != nil {
		t.Fatalf("BuildSupportDocument: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/support_document_golden.xml"

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
