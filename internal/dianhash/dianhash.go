// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
//
// Package dianhash contains the field-concatenation logic shared by the CUFE (package cufe)
// and CUDE (package cude) formulas — Technical Annex 1.9, sections 11.2 and 11.4. It is
// "internal" because it exists only so those two packages don't duplicate the same formula —
// it is not a public API of the module.
package dianhash

import "github.com/diegofxm/cofacture/domain"

// Seed builds the input string for the SHA-384 hash shared by CUFE and CUDE: same fields,
// same order. The only thing that differs between the two is lastSeedComponent (the numbering
// range's technical key for CUFE, the software PIN for CUDE) — the caller decides which one
// it is.
func Seed(doc domain.Invoice, lastSeedComponent string) string {
	var ivaCents, incCents, icaCents int64
	for _, t := range doc.HeaderTaxes {
		switch t.TypeCode {
		case "01":
			ivaCents += t.TaxAmountCents
		case "04":
			incCents += t.TaxAmountCents
		case "03":
			icaCents += t.TaxAmountCents
		}
	}

	return doc.Prefix + doc.Number +
		doc.IssueDate +
		doc.IssueTime +
		domain.FormatCents(doc.Totals.LineExtensionCents) +
		"01" + domain.FormatCents(ivaCents) +
		"04" + domain.FormatCents(incCents) +
		"03" + domain.FormatCents(icaCents) +
		domain.FormatCents(doc.Totals.PayableCents) +
		doc.Supplier.Identification.Number +
		doc.Customer.Identification.Number +
		lastSeedComponent +
		doc.EnvironmentCode
}
