// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"strconv"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
	ubl "github.com/diegofxm/cofacture/xml"
)

// BuildDebitNote builds the XML tree of a Debit Note. Nearly identical to BuildCreditNote —
// the real differences are: there's no equivalent to cbc:CreditNoteTypeCode, the totals block
// is named cac:RequestedMonetaryTotal instead of cac:LegalMonetaryTotal, and the customer does
// carry PhysicalLocation (unlike Invoice/CreditNote, where it doesn't apply).
func BuildDebitNote(dn domain.DebitNote) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8" standalone="no"`)

	root := doc.CreateElement("DebitNote")
	root.CreateAttr("xmlns", ubl.NSDebitNote)
	root.CreateAttr("xmlns:cac", ubl.NSCac)
	root.CreateAttr("xmlns:cbc", ubl.NSCbc)
	root.CreateAttr("xmlns:ds", ubl.NSDs)
	root.CreateAttr("xmlns:ext", ubl.NSExt)
	root.CreateAttr("xmlns:sts", ubl.NSSts)
	root.CreateAttr("xmlns:xades", ubl.NSXades)
	root.CreateAttr("xmlns:xades141", ubl.NSXades141)
	root.CreateAttr("xmlns:xsi", ubl.NSXsi)
	root.CreateAttr("xsi:schemaLocation", ubl.NSDebitNote+" http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/UBL-DebitNote-2.1.xsd")

	appendUBLExtensions(root, dn.Invoice)

	root.CreateElement("cbc:UBLVersionID").SetText("UBL 2.1")
	root.CreateElement("cbc:CustomizationID").SetText(dn.OperationTypeCode)
	root.CreateElement("cbc:ProfileID").SetText(dn.ProfileID)
	root.CreateElement("cbc:ProfileExecutionID").SetText(dn.EnvironmentCode)
	root.CreateElement("cbc:ID").SetText(dn.Prefix + dn.Number)

	uuid := root.CreateElement("cbc:UUID")
	uuid.CreateAttr("schemeID", dn.EnvironmentCode)
	uuid.CreateAttr("schemeName", dn.HashType)
	uuid.SetText(dn.CUFE) // this note's CUDE, see the note on domain.DebitNote

	root.CreateElement("cbc:IssueDate").SetText(dn.IssueDate)
	root.CreateElement("cbc:IssueTime").SetText(dn.IssueTime)
	if dn.Note != "" {
		root.CreateElement("cbc:Note").SetText(dn.Note)
	}

	currency := root.CreateElement("cbc:DocumentCurrencyCode")
	currency.CreateAttr("listAgencyID", "6")
	currency.CreateAttr("listAgencyName", "United Nations Economic Commission for Europe")
	currency.CreateAttr("listID", "ISO 4217 Alpha")
	currency.SetText(dn.CurrencyCode)

	root.CreateElement("cbc:LineCountNumeric").SetText(strconv.Itoa(len(dn.Lines)))

	if dn.DiscrepancyResponse != nil {
		appendDiscrepancyResponse(root, *dn.DiscrepancyResponse)
	}
	appendBillingReference(root, "InvoiceDocumentReference", dn.BillingReference)

	appendAccountingParty(root, "AccountingSupplierParty", dn.Supplier, true, dn.Prefix, true)
	appendAccountingParty(root, "AccountingCustomerParty", dn.Customer, false, "", true)

	for _, pm := range dn.PaymentMeans {
		appendPaymentMean(root, pm)
	}

	appendTaxTotal(root, dn.HeaderTaxes, dn.CurrencyCode)
	appendMonetaryTotal(root, "RequestedMonetaryTotal", dn.Totals, dn.CurrencyCode, dn.DocumentTypeCode)

	for i, line := range dn.Lines {
		appendDocumentLine(root, "DebitNoteLine", "DebitedQuantity", i+1, line, dn.CurrencyCode, dn.DocumentTypeCode, dn.Supplier.Identification, "")
	}

	return doc, nil
}
