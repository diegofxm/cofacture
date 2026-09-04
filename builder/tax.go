// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
)

// appendWithholdingTaxTotal adds a cac:WithholdingTaxTotal for each withholding on the Support
// Document (InvoiceTypeCode "05"). Unlike cac:TaxTotal, where every subtype goes under the
// same element, each withholding generates its own cac:WithholdingTaxTotal with exactly one
// cac:TaxSubtotal.
func appendWithholdingTaxTotal(parent *etree.Element, taxes []domain.Tax, currency string) {
	for _, t := range taxes {
		wtt := parent.CreateElement("cac:WithholdingTaxTotal")

		amt := wtt.CreateElement("cbc:TaxAmount")
		amt.CreateAttr("currencyID", currency)
		amt.SetText(formatAmount(t.TaxAmountCents))

		sub := wtt.CreateElement("cac:TaxSubtotal")

		taxable := sub.CreateElement("cbc:TaxableAmount")
		taxable.CreateAttr("currencyID", currency)
		taxable.SetText(formatAmount(t.TaxableAmountCents))

		taxAmt := sub.CreateElement("cbc:TaxAmount")
		taxAmt.CreateAttr("currencyID", currency)
		taxAmt.SetText(formatAmount(t.TaxAmountCents))

		category := sub.CreateElement("cac:TaxCategory")
		category.CreateElement("cbc:Percent").SetText(formatPercent(t.Percent))
		category.CreateElement("cbc:TaxExemptionReason").SetText("01")
		scheme := category.CreateElement("cac:TaxScheme")
		scheme.CreateElement("cbc:ID").SetText(t.TypeCode)
		scheme.CreateElement("cbc:Name").SetText(t.TypeName)
	}
}

// appendTaxTotal adds cac:TaxTotal (header or line level). Adds nothing if taxes is empty,
// matching the technical annex's requirement that TaxTotal is optional when no tax applies.
func appendTaxTotal(parent *etree.Element, taxes []domain.Tax, currency string) {
	if len(taxes) == 0 {
		return
	}

	var totalCents int64
	for _, t := range taxes {
		totalCents += t.TaxAmountCents
	}

	taxTotal := parent.CreateElement("cac:TaxTotal")
	amt := taxTotal.CreateElement("cbc:TaxAmount")
	amt.CreateAttr("currencyID", currency)
	amt.SetText(formatAmount(totalCents))

	for _, t := range taxes {
		sub := taxTotal.CreateElement("cac:TaxSubtotal")

		taxable := sub.CreateElement("cbc:TaxableAmount")
		taxable.CreateAttr("currencyID", currency)
		taxable.SetText(formatAmount(t.TaxableAmountCents))

		taxAmt := sub.CreateElement("cbc:TaxAmount")
		taxAmt.CreateAttr("currencyID", currency)
		taxAmt.SetText(formatAmount(t.TaxAmountCents))

		category := sub.CreateElement("cac:TaxCategory")
		category.CreateElement("cbc:Percent").SetText(formatPercent(t.Percent))
		scheme := category.CreateElement("cac:TaxScheme")
		scheme.CreateElement("cbc:ID").SetText(t.TypeCode)
		scheme.CreateElement("cbc:Name").SetText(t.TypeName)
	}
}
