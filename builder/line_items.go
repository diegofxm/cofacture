// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strconv"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
)

// documentTypeCodesUsingMandante is the closed set of document types whose Item carries
// InformationContentProviderParty/PowerOfAttorney/AgentParty (the "Mandante"/principal's
// identification) instead of StandardItemIdentification: Credit Note ("91"), Debit Note ("92"),
// and their Documento Equivalente Electrónico Adjustment Note equivalents ("93"/"94", Technical
// Annex Documento Equivalente Electrónico V1.0 section 16.3). Every other document type —
// Invoice, Support Document, Adjustment Note to the Support Document, and every Documento
// Equivalente Electrónico type (20/25/27/30/35/40/45/50/55/60, same annex) — uses
// StandardItemIdentification instead. This used to be the other way around (an allowlist of
// "01"/"05"/"95" defaulting everything else to Mandante), which silently broke the first time a
// new primary-document type code was introduced (Documento Equivalente Electrónico "20" fell
// into the Mandante branch by accident — caught by builder/pos_test.go, not by inspection).
var documentTypeCodesUsingMandante = map[string]bool{"91": true, "92": true, "93": true, "94": true}

// appendDocumentLine adds a detail line — shared by Invoice ("InvoiceLine"/
// "InvoicedQuantity"), CreditNote ("CreditNoteLine"/"CreditedQuantity") and DebitNote
// ("DebitNoteLine"/"DebitedQuantity"); node/nodeQty select which.
//
// See documentTypeCodesUsingMandante for which of the two the Item element carries; verified
// against the technical annex (group FBA0x/CBA0x) and against a reference generator in
// real-world use.
// invoicePeriodStartDate: when not empty, adds cac:InvoicePeriod to the line — required on
// every line of the Support Document (DSFC01). Empty for Invoice/CreditNote/DebitNote.
func appendDocumentLine(parent *etree.Element, node, nodeQty string, index int, line domain.Line, currency, documentTypeCode string, mandanteID domain.Identification, invoicePeriodStartDate string) {
	el := parent.CreateElement("cac:" + node)
	el.CreateElement("cbc:ID").SetText(strconv.Itoa(index))

	qty := el.CreateElement("cbc:" + nodeQty)
	qty.CreateAttr("unitCode", line.UnitCode)
	qty.SetText(formatQuantity(line.Quantity))

	lineExt := el.CreateElement("cbc:LineExtensionAmount")
	lineExt.CreateAttr("currencyID", currency)
	lineExt.SetText(formatAmount(line.LineExtensionCents))

	if documentTypeCode != "92" {
		el.CreateElement("cbc:FreeOfChargeIndicator").SetText(formatBool(line.FreeOfCharge))
	}

	if line.FreeOfCharge && line.ReferencePrice != nil {
		altPrice := el.CreateElement("cac:PricingReference").CreateElement("cac:AlternativeConditionPrice")
		price := altPrice.CreateElement("cbc:PriceAmount")
		price.CreateAttr("currencyID", currency)
		price.SetText(formatAmount(line.ReferencePrice.PriceAmountCents))
		altPrice.CreateElement("cbc:PriceTypeCode").SetText(line.ReferencePrice.TypeCode)
	}

	if invoicePeriodStartDate != "" {
		ip := el.CreateElement("cac:InvoicePeriod")
		ip.CreateElement("cbc:StartDate").SetText(invoicePeriodStartDate)
		ip.CreateElement("cbc:DescriptionCode").SetText("1")
		ip.CreateElement("cbc:Description").SetText("Por operación")
	}

	appendTaxTotal(el, line.Taxes, currency)

	item := el.CreateElement("cac:Item")
	item.CreateElement("cbc:Description").SetText(line.Description)
	if documentTypeCodesUsingMandante[documentTypeCode] {
		agentID := item.CreateElement("cac:InformationContentProviderParty").
			CreateElement("cac:PowerOfAttorney").
			CreateElement("cac:AgentParty").
			CreateElement("cac:PartyIdentification").
			CreateElement("cbc:ID")
		setIdentificationAttrs(agentID, mandanteID)
		agentID.SetText(mandanteID.Number)
	} else {
		sii := item.CreateElement("cac:StandardItemIdentification").CreateElement("cbc:ID")
		sii.CreateAttr("schemeID", line.ItemTypeCode)
		sii.CreateAttr("schemeName", line.ItemTypeName)
		// schemeAgencyID is only added when provided — code "999" (the taxpayer's own
		// standard, table 13.3.5 of the Technical Annex) explicitly requires this attribute be
		// omitted entirely, not even left empty (same rule as cac:Country in
		// appendAddressFields).
		if line.ItemTypeAgencyID != "" {
			sii.CreateAttr("schemeAgencyID", line.ItemTypeAgencyID)
		}
		sii.SetText(line.ItemCode)
	}

	price := el.CreateElement("cac:Price")
	priceAmt := price.CreateElement("cbc:PriceAmount")
	priceAmt.CreateAttr("currencyID", currency)
	priceAmt.SetText(formatAmount(line.UnitPriceCents))
	baseQty := price.CreateElement("cbc:BaseQuantity")
	baseQty.CreateAttr("unitCode", line.UnitCode)
	baseQty.SetText(formatQuantity(line.Quantity))
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
