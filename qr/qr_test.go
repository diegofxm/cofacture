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
		t.Errorf("expected Amb:1 in the content, got:\n%s", got)
	}
}

// TestSupportDocumentContent_NoVATFallsBackToFirstTax confirms that when HeaderTaxes has
// entries but none is VAT ("01"), CodImp/ValImp fall back to the first tax in the slice instead
// of the "01"/"0.00" default.
func TestSupportDocumentContent_NoVATFallsBackToFirstTax(t *testing.T) {
	inv := domain.Invoice{
		Number:          "1",
		IssueDate:       "2024-01-01",
		IssueTime:       "00:00:00-05:00",
		EnvironmentCode: "2",
		HeaderTaxes: []domain.Tax{
			{TypeCode: "04", TaxAmountCents: 500_00}, // ICA, not VAT
		},
		Totals:   domain.Totals{LineExtensionCents: 100_000, PayableCents: 105_000},
		Supplier: domain.Party{Identification: domain.Identification{Number: "111"}},
		Customer: domain.Party{Identification: domain.Identification{Number: "222"}},
	}
	got := SupportDocumentContent(inv, "cuds", "pin")
	if !strings.Contains(got, "CodImp=04") || !strings.Contains(got, "ValImp=500.00") {
		t.Errorf("expected fallback to the first tax (04/500.00), got:\n%s", got)
	}
}

// TestSupportDocumentContent_NoTaxesDefaultsToZero confirms that with zero HeaderTaxes,
// CodImp/ValImp default to "01"/"0.00" instead of being left empty.
func TestSupportDocumentContent_NoTaxesDefaultsToZero(t *testing.T) {
	inv := domain.Invoice{
		Number:          "1",
		IssueDate:       "2024-01-01",
		IssueTime:       "00:00:00-05:00",
		EnvironmentCode: "2",
		Totals:          domain.Totals{LineExtensionCents: 100_000, PayableCents: 100_000},
		Supplier:        domain.Party{Identification: domain.Identification{Number: "111"}},
		Customer:        domain.Party{Identification: domain.Identification{Number: "222"}},
	}
	got := SupportDocumentContent(inv, "cuds", "pin")
	if !strings.Contains(got, "CodImp=01") || !strings.Contains(got, "ValImp=0.00") {
		t.Errorf("expected the default CodImp=01/ValImp=0.00 with no HeaderTaxes, got:\n%s", got)
	}
}

func TestSupportDocumentURL(t *testing.T) {
	cases := []struct {
		env, want string
	}{
		{"1", "https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=cuds123"},
		{"2", "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=cuds123"},
	}
	for _, c := range cases {
		if got := SupportDocumentURL(c.env, "cuds123"); got != c.want {
			t.Errorf("SupportDocumentURL(%q, ...) = %q, want %q", c.env, got, c.want)
		}
	}
}

func TestAdjustmentNoteContent(t *testing.T) {
	inv := domain.Invoice{
		Prefix:          "NAP",
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
		// Roles reversed, same as the Support Document it adjusts: Supplier = SNO, Customer = ABS.
		Supplier: domain.Party{Identification: domain.Identification{Number: "1020304050"}},
		Customer: domain.Party{Identification: domain.Identification{Number: "900123456"}},
	}

	const cuds = "907e4444decc9e59c160a2fb3b6659b33dc5b632a5008922b9a62f83f757b1c448e47f5867f2b50dbdb96f48c7681168"
	const pin = "12345"

	got := AdjustmentNoteContent(inv, cuds, pin)

	required := []string{
		"N°NotaAjuste=NAP1",
		"Fecha=2024-03-15",
		"Hora=09:30:00-05:00",
		"ValNA=100000.00",
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
			t.Errorf("AdjustmentNoteContent() does not contain %q\n--- got ---\n%s", r, got)
		}
	}

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	wantURL := "URL=https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=" + cuds
	if lines[len(lines)-1] != wantURL {
		t.Errorf("last line = %q, want %q", lines[len(lines)-1], wantURL)
	}
}

// TestAdjustmentNoteContent_Produccion verifies that environment "1" uses the production domain.
func TestAdjustmentNoteContent_Produccion(t *testing.T) {
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
	got := AdjustmentNoteContent(inv, cuds, "000")

	wantURL := "URL=https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=abc"
	if !strings.Contains(got, wantURL) {
		t.Errorf("expected the production URL in the content, got:\n%s", got)
	}
	if !strings.Contains(got, "Amb:1") {
		t.Errorf("expected Amb:1 in the content, got:\n%s", got)
	}
}

// TestAdjustmentNoteContent_NoTaxesDefaultsToZero mirrors
// TestSupportDocumentContent_NoTaxesDefaultsToZero for AdjustmentNoteContent's own (duplicated)
// fallback logic.
func TestAdjustmentNoteContent_NoTaxesDefaultsToZero(t *testing.T) {
	inv := domain.Invoice{
		Number:          "1",
		IssueDate:       "2024-01-01",
		IssueTime:       "00:00:00-05:00",
		EnvironmentCode: "2",
		Totals:          domain.Totals{LineExtensionCents: 100_000, PayableCents: 100_000},
		Supplier:        domain.Party{Identification: domain.Identification{Number: "111"}},
		Customer:        domain.Party{Identification: domain.Identification{Number: "222"}},
	}
	got := AdjustmentNoteContent(inv, "cuds", "pin")
	if !strings.Contains(got, "CodImp=01") || !strings.Contains(got, "ValImp=0.00") {
		t.Errorf("expected the default CodImp=01/ValImp=0.00 with no HeaderTaxes, got:\n%s", got)
	}
}
