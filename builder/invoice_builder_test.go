// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"flag"
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

// update regenerates testdata/invoice_golden.xml from the builder's current output.
// Usage: go test ./builder/... -run TestBuildInvoice_Golden -update
// After regenerating, manually review the diff against the reference template before
// trusting the new version of the golden file.
var update = flag.Bool("update", false, "regenera el golden file con la salida actual")

func ptr(s string) *string { return &s }

func sampleInvoice() domain.Invoice {
	return domain.Invoice{
		ProfileID:         "DIAN 2.1",
		EnvironmentCode:   "2",
		OperationTypeCode: "10",
		DocumentTypeCode:  "01",
		HashType:          "CUFE-SHA384",

		Prefix: "SETP",
		Number: "1",

		IssueDate: "2024-01-20",
		IssueTime: "10:00:00-05:00",

		CurrencyCode: "COP",

		Supplier: domain.Party{
			EntityTypeCode: "2",
			Identification: domain.Identification{
				Number:           "900123456",
				TypeCode:         "31",
				VerificationCode: "3",
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
			TaxRegimeCode:               "48",
			LiabilityCodes:              []string{"O-13", "O-15", "O-23"},
			IndustryClassificationCodes: []string{"4661", "4669"},
			TaxSchemeCode:               "01",
			TaxSchemeName:               "IVA",
			Phone:                       "3001234567",
			Email:                       "contacto@miempresa.com",
			MerchantRegistrationNumber:  ptr("12345"),
		},

		Customer: domain.Party{
			EntityTypeCode: "1",
			Identification: domain.Identification{
				Number:   "1234567890",
				TypeCode: "13",
			},
			Name:           "Juan Pérez",
			TaxRegimeCode:  "49",
			LiabilityCodes: []string{"O-47"},
			TaxSchemeCode:  "01",
			TaxSchemeName:  "IVA",
			Phone:          "3009876543",
			Email:          "juan.perez@example.com",
		},

		PaymentMeans: []domain.PaymentMean{
			{Code: "1", PaymentMethodCode: "10"},
		},

		HeaderTaxes: []domain.Tax{
			{TaxableAmountCents: 126050420, TaxAmountCents: 23949580, Percent: 19, TypeCode: "01", TypeName: "IVA"},
		},

		Totals: domain.Totals{
			LineExtensionCents: 126050420,
			TaxExclusiveCents:  126050420,
			TaxInclusiveCents:  150000000,
			PayableCents:       150000000,
		},

		Lines: []domain.Line{
			{
				Description:        "Laptop Dell Inspiron 15, 8GB RAM, 256GB SSD",
				Quantity:           1,
				UnitCode:           "94",
				LineExtensionCents: 126050420,
				UnitPriceCents:     126050420,
				ItemCode:           "PROD001",
				ItemTypeCode:       "999",
				ItemTypeName:       "Estándar de adopción del contribuyente",
				// ItemTypeAgencyID intentionally empty — row "999" of table 13.3.5 in the
				// Technical Annex requires that @schemeAgencyID not be used at all (see
				// section 9.45), not that it be sent with a value of "0".
				Taxes: []domain.Tax{
					{TaxableAmountCents: 126050420, TaxAmountCents: 23949580, Percent: 19, TypeCode: "01", TypeName: "IVA"},
				},
			},
		},

		NumberingRange: domain.NumberingRange{
			AuthorizedCode: "18760000001",
			Prefix:         "SETP",
			StartNumber:    "1",
			EndNumber:      "5000",
			StartDate:      "2024-01-15",
			EndDate:        "2026-01-15",
		},

		SoftwareProvider: domain.SoftwareProvider{
			ProviderIdentification: domain.Identification{
				Number:           "900123456",
				TypeCode:         "31",
				VerificationCode: "3",
			},
			SoftwareID: "fac4203d-2451-4806-8a3e-000000000001",
		},
	}
}

func TestBuildInvoice_Golden(t *testing.T) {
	doc, err := BuildInvoice(sampleInvoice())
	if err != nil {
		t.Fatalf("BuildInvoice: %v", err)
	}
	doc.Indent(2)
	got, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	const goldenPath = "testdata/invoice_golden.xml"

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
