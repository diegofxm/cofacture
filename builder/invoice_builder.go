// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
//
// Package builder assembles UBL 2.1 + DIAN extension documents from the domain models. It
// does not sign, does not compute CUFE/QR, and knows nothing about HTTP or a database — it
// only assembles the XML.
package builder

import (
	"strconv"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
	ubl "github.com/diegofxm/cofacture/xml"
)

// BuildInvoice builds the XML tree of an Electronic Sales Invoice.
//
// The resulting document still has no real CUFE, SoftwareSecurityCode or QRURL if the Invoice
// it received had them empty — those values are computed in later pipeline steps (cufe, qr)
// from this same model and injected before signing.
func BuildInvoice(inv domain.Invoice) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8" standalone="no"`)

	root := doc.CreateElement("Invoice")
	root.CreateAttr("xmlns", ubl.NSInvoiceDefault)
	root.CreateAttr("xmlns:cac", ubl.NSCac)
	root.CreateAttr("xmlns:cbc", ubl.NSCbc)
	root.CreateAttr("xmlns:ds", ubl.NSDs)
	root.CreateAttr("xmlns:ext", ubl.NSExt)
	root.CreateAttr("xmlns:sts", ubl.NSSts)
	root.CreateAttr("xmlns:xades", ubl.NSXades)
	root.CreateAttr("xmlns:xades141", ubl.NSXades141)
	root.CreateAttr("xmlns:xsi", ubl.NSXsi)
	root.CreateAttr("xsi:schemaLocation", ubl.SchemaLocationInvoice)

	appendUBLExtensions(root, inv)

	root.CreateElement("cbc:UBLVersionID").SetText("UBL 2.1")
	root.CreateElement("cbc:CustomizationID").SetText(inv.OperationTypeCode)
	root.CreateElement("cbc:ProfileID").SetText(inv.ProfileID)
	root.CreateElement("cbc:ProfileExecutionID").SetText(inv.EnvironmentCode)
	root.CreateElement("cbc:ID").SetText(inv.Prefix + inv.Number)

	uuid := root.CreateElement("cbc:UUID")
	uuid.CreateAttr("schemeID", inv.EnvironmentCode)
	uuid.CreateAttr("schemeName", inv.HashType)
	uuid.SetText(inv.CUFE)

	root.CreateElement("cbc:IssueDate").SetText(inv.IssueDate)
	root.CreateElement("cbc:IssueTime").SetText(inv.IssueTime)
	if inv.DueDate != "" {
		root.CreateElement("cbc:DueDate").SetText(inv.DueDate)
	}
	root.CreateElement("cbc:InvoiceTypeCode").SetText(inv.DocumentTypeCode)
	if inv.Note != "" {
		root.CreateElement("cbc:Note").SetText(inv.Note)
	}

	currency := root.CreateElement("cbc:DocumentCurrencyCode")
	currency.CreateAttr("listAgencyID", "6")
	currency.CreateAttr("listAgencyName", "United Nations Economic Commission for Europe")
	currency.CreateAttr("listID", "ISO 4217 Alpha")
	currency.SetText(inv.CurrencyCode)

	root.CreateElement("cbc:LineCountNumeric").SetText(strconv.Itoa(len(inv.Lines)))

	if inv.OrderReferenceNumber != "" {
		root.CreateElement("cac:OrderReference").CreateElement("cbc:ID").SetText(inv.OrderReferenceNumber)
	}

	appendAccountingParty(root, "AccountingSupplierParty", inv.Supplier, true, inv.Prefix, true)
	appendAccountingParty(root, "AccountingCustomerParty", inv.Customer, false, "", false)

	for _, pm := range inv.PaymentMeans {
		appendPaymentMean(root, pm)
	}

	appendTaxTotal(root, inv.HeaderTaxes, inv.CurrencyCode)
	appendMonetaryTotal(root, "LegalMonetaryTotal", inv.Totals, inv.CurrencyCode, inv.DocumentTypeCode)

	for i, line := range inv.Lines {
		// mandanteID does not apply to Invoice (only notes use it); passed empty.
		appendDocumentLine(root, "InvoiceLine", "InvoicedQuantity", i+1, line, inv.CurrencyCode, inv.DocumentTypeCode, domain.Identification{}, "")
	}

	return doc, nil
}
