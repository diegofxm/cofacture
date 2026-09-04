// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package dianhash

import (
	"strings"
	"testing"

	"github.com/diegofxm/cofacture/domain"
)

// TestSeed_AnexoTecnicoExample uses the same official worked example from section 11.2.1 of the
// Technical Annex 1.9 that cufe.TestCompute_AnexoTecnicoExample hashes — but asserts the exact
// concatenated seed string itself, not just the resulting hash. A wrong field order or a wrong
// separator would still be caught by the hash test, but only as an opaque mismatch; asserting
// the string directly here makes a future regression immediately diagnosable.
func TestSeed_AnexoTecnicoExample(t *testing.T) {
	inv := domain.Invoice{
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
	const want = "323200000129" + "2019-01-16" + "10:53:10-05:00" + "1500000.00" +
		"01" + "285000.00" + "04" + "0.00" + "03" + "0.00" +
		"1785000.00" + "700085371" + "800199436" + technicalKey + "1"

	if got := Seed(inv, technicalKey); got != want {
		t.Errorf("Seed() = %q, want %q", got, want)
	}
}

// TestSeed_PrefixNumberConcatenation confirms NumFac is Prefix+Number with no separator —
// the DIAN worked example above leaves Prefix empty, so it alone doesn't prove concatenation
// order.
func TestSeed_PrefixNumberConcatenation(t *testing.T) {
	inv := domain.Invoice{Prefix: "SETP", Number: "990000001"}
	got := Seed(inv, "key")
	const wantPrefix = "SETP990000001"
	if len(got) < len(wantPrefix) || got[:len(wantPrefix)] != wantPrefix {
		t.Errorf("Seed() does not start with %q (Prefix+Number): %q", wantPrefix, got)
	}
}

// TestSeed_AccumulatesMultipleTaxesOfSameType confirms two HeaderTaxes entries with the same
// TypeCode sum into the same slot instead of only the first or last one counting.
func TestSeed_AccumulatesMultipleTaxesOfSameType(t *testing.T) {
	inv := domain.Invoice{
		HeaderTaxes: []domain.Tax{
			{TypeCode: "01", TaxAmountCents: 100_00},
			{TypeCode: "01", TaxAmountCents: 50_00},
		},
	}
	got := Seed(inv, "key")
	const wantIVA = "01150.00" // 100.00 + 50.00
	if !strings.Contains(got, wantIVA) {
		t.Errorf("Seed() = %q, want it to contain the summed IVA slot %q", got, wantIVA)
	}
}

// TestSeed_AllThreeTaxSlots confirms IVA/INC/ICA each land in their own fixed slot, in order,
// regardless of the order HeaderTaxes lists them in.
func TestSeed_AllThreeTaxSlots(t *testing.T) {
	inv := domain.Invoice{
		HeaderTaxes: []domain.Tax{
			{TypeCode: "03", TaxAmountCents: 300_00}, // ICA, listed first
			{TypeCode: "01", TaxAmountCents: 100_00}, // IVA
			{TypeCode: "04", TaxAmountCents: 200_00}, // INC
		},
	}
	got := Seed(inv, "key")
	const wantSlots = "01100.00" + "04200.00" + "03300.00" // fixed order: IVA, INC, ICA
	if !strings.Contains(got, wantSlots) {
		t.Errorf("Seed() = %q, want it to contain the fixed-order tax slots %q", got, wantSlots)
	}
}

// TestSeed_IgnoresOtherTaxTypes confirms a tax type outside IVA/INC/ICA (e.g. withholdings,
// which belong in a separate WithholdingTaxTotal, not HeaderTaxes) contributes to none of the
// three fixed slots.
func TestSeed_IgnoresOtherTaxTypes(t *testing.T) {
	inv := domain.Invoice{
		HeaderTaxes: []domain.Tax{
			{TypeCode: "06", TaxAmountCents: 999_00}, // ReteRenta — not part of the CUFE/CUDE formula
		},
	}
	got := Seed(inv, "key")
	const wantSlots = "010.00" + "040.00" + "030.00" // all three slots stay at zero
	if !strings.Contains(got, wantSlots) {
		t.Errorf("Seed() = %q, want the untouched slots %q (type 06 must not count)", got, wantSlots)
	}
}
