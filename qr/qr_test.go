// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package qr

import (
	"strings"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

func TestURL(t *testing.T) {
	cases := []struct {
		env, want string
	}{
		{"1", "https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=abc123"},
		{"2", "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=abc123"},
	}
	for _, c := range cases {
		if got := URL(c.env, "abc123"); got != c.want {
			t.Errorf("URL(%q, ...) = %q, want %q", c.env, got, c.want)
		}
	}
}

func TestSupportDocumentContent(t *testing.T) {
	inv := domain.Invoice{
		Prefix:          "DS",
		Number:          "1",
		IssueDate:       "2024-03-15",
		IssueTime:       "09:30:00-05:00",
		EnvironmentCode: "2",
		HeaderTaxes: []domain.Tax{
			{TypeCode: "01", TaxAmountCents: 1_900_000},
		},
		Totals: domain.Totals{
			LineExtensionCents: 10_000_000,
			PayableCents:       11_900_000,
		},
		// DS: Supplier = non-obligated third party (supplier), Customer = issuer.
		Supplier: domain.Party{Identification: domain.Identification{Number: "1020304050"}},
		Customer: domain.Party{Identification: domain.Identification{Number: "900123456"}},
	}

	const cuds = "907e4444decc9e59c160a2fb3b6659b33dc5b632a5008922b9a62f83f757b1c448e47f5867f2b50dbdb96f48c7681168"
	const pin = "12345"

	got := SupportDocumentContent(inv, cuds, pin)

	// The 12 mandatory tags of the Support Document QR (Annex 1.9, section 11.7.1).
	required := []string{
		"N°DocSoporte=DS1",
		"Fecha=2024-03-15",
		"Hora=09:30:00-05:00",
		"ValDS=100000.00",
		"CodImp=01",
		"ValImp=19000.00",
		"ValTot=119000.00",
		"NumSNO=1020304050",
		"NITABS=900123456",
		"PIN:12345",
		"Amb:2",
		"CUDS=" + cuds,
	}
	for _, r := range required {
		if !strings.Contains(got, r) {
			t.Errorf("SupportDocumentContent() no contiene %q\n--- got ---\n%s", r, got)
		}
	}

	// The last line must be "URL=<searchqr>" — same endpoint as FE (FindDocument does not redirect).
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	wantURL := "URL=https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=" + cuds
	if lines[len(lines)-1] != wantURL {
		t.Errorf("last line = %q, want %q", lines[len(lines)-1], wantURL)
	}
}

// TestSupportDocumentContent_Produccion verifies that environment "1" uses the production domain.
func TestSupportDocumentContent_Produccion(t *testing.T) {
	inv := domain.Invoice{
		Number:          "99",
		IssueDate:       "2024-01-01",
		IssueTime:       "00:00:00-05:00",
		EnvironmentCode: "1",
		Totals:          domain.Totals{LineExtensionCents: 100_000, PayableCents: 100_000},
		Supplier:        domain.Party{Identification: domain.Identification{Number: "111"}},
		Customer:        domain.Party{Identification: domain.Identification{Number: "222"}},
	}
	const cuds = "abc"
	got := SupportDocumentContent(inv, cuds, "000")

	wantURL := "URL=https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=abc"
	if !strings.Contains(got, wantURL) {
		t.Errorf("expected the production URL in the content, got:\n%s", got)
	}
	if !strings.Contains(got, "Amb:1") {
		t.Errorf("esperaba Amb:1 en el contenido, got:\n%s", got)
	}
}
