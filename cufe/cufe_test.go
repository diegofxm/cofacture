// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package cufe

import (
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

// TestCompute_AnexoTecnicoExample uses the official example from section 11.2.1 of the
// Technical Annex 1.9 (DIAN Resolution 000165/2023) — same input values and same expected
// hash that DIAN publishes, not a made-up case.
func TestCompute_AnexoTecnicoExample(t *testing.T) {
	inv := domain.Document{
		Number:          "323200000129",
		IssueDate:       "2019-01-16",
		IssueTime:       "10:53:10-05:00",
		EnvironmentCode: "1",
		Totals: domain.Totals{
			LineExtensionCents: 150_000_000, // 1500000.00
			PayableCents:       178_500_000, // 1785000.00
		},
		HeaderTaxes: []domain.Tax{
			{TypeCode: "01", TaxAmountCents: 28_500_000}, // IVA 285000.00
		},
		Supplier: domain.Party{Identification: domain.Identification{Number: "700085371"}},
		Customer: domain.Party{Identification: domain.Identification{Number: "800199436"}},
	}

	const technicalKey = "693ff6f2a553c3646a063436fd4dd9ded0311471"
	const want = "8bb918b19ba22a694f1da11c643b5e9de39adf60311cf179179e9b33381030bcd4c3c3f156c506ed5908f9276f5bd9b4"

	got := Compute(inv, technicalKey)
	if got != want {
		t.Errorf("Compute() = %s, want %s", got, want)
	}
}
