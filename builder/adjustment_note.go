// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strconv"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
	ubl "github.com/diegofxm/cofacture/xml"
)

// BuildAdjustmentNote builds the XML tree of an Adjustment Note to the Support Document
// (InvoiceTypeCode "95"). Uses CreditNote as the root element (same as a regular Credit
// Note/91) with the CreditNote-2 namespace — verified against DIAN's official
// NotaDeAjuste.xml example. Roles are inverted just like the Support Document: Supplier = SNO
// (non-obligated third party), Customer = ABS.
func BuildAdjustmentNote(an domain.AdjustmentNote) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8" standalone="no"`)

	root := doc.CreateElement("CreditNote")
	root.CreateAttr("xmlns", ubl.NSCreditNote)
	root.CreateAttr("xmlns:cac", ubl.NSCac)
	root.CreateAttr("xmlns:cbc", ubl.NSCbc)
	root.CreateAttr("xmlns:ds", ubl.NSDs)
	root.CreateAttr("xmlns:ext", ubl.NSExt)
	root.CreateAttr("xmlns:sts", ubl.NSSts)
	root.CreateAttr("xmlns:xades", ubl.NSXades)
	root.CreateAttr("xmlns:xades141", ubl.NSXades141)
	root.CreateAttr("xmlns:xsi", ubl.NSXsi)
	root.CreateAttr("xsi:schemaLocation", ubl.NSCreditNote+" http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/UBL-CreditNote-2.1.xsd")

	// appendUBLExtensions handles InvoiceControl for type "05"/"95" (Support Document family).
	appendUBLExtensions(root, an.Invoice)

	root.CreateElement("cbc:UBLVersionID").SetText("UBL 2.1")
	root.CreateElement("cbc:CustomizationID").SetText(an.OperationTypeCode)
	root.CreateElement("cbc:ProfileID").SetText(an.ProfileID)
	root.CreateElement("cbc:ProfileExecutionID").SetText(an.EnvironmentCode)
	root.CreateElement("cbc:ID").SetText(an.Prefix + an.Number)

	uuid := root.CreateElement("cbc:UUID")
	uuid.CreateAttr("schemeID", an.EnvironmentCode)
	uuid.CreateAttr("schemeName", an.HashType) // "CUDS-SHA384"
	uuid.SetText(an.CUFE)

	root.CreateElement("cbc:IssueDate").SetText(an.IssueDate)
	root.CreateElement("cbc:IssueTime").SetText(an.IssueTime)
	root.CreateElement("cbc:CreditNoteTypeCode").SetText(an.DocumentTypeCode) // "95"
	if an.Note != "" {
		root.CreateElement("cbc:Note").SetText(an.Note)
	}

	// N/A: DocumentCurrencyCode with no list attributes (same as the Support Document — DSFC03).
	root.CreateElement("cbc:DocumentCurrencyCode").SetText(an.CurrencyCode)

	root.CreateElement("cbc:LineCountNumeric").SetText(strconv.Itoa(len(an.Lines)))

	// DiscrepancyResponse goes before BillingReference (same order as Credit/Debit Note).
	if an.DiscrepancyResponse != nil {
		appendDiscrepancyResponse(root, *an.DiscrepancyResponse)
	}

	// BillingReference points to the original Support Document — the UUID carries
	// schemeName="CUDS-SHA384".
	appendNABillingReference(root, an.BillingReference)

	// Roles inverted just like the Support Document: Supplier = non-obligated third party,
	// Customer = ABS.
	appendDSSupplierParty(root, an.Supplier)
	appendDSCustomerParty(root, an.Customer)

	for _, pm := range an.PaymentMeans {
		appendPaymentMean(root, pm)
	}

	appendTaxTotal(root, an.HeaderTaxes, an.CurrencyCode)
	appendWithholdingTaxTotal(root, an.WithholdingTaxes, an.CurrencyCode)
	appendMonetaryTotal(root, "LegalMonetaryTotal", an.Totals, an.CurrencyCode, an.DocumentTypeCode)

	linePeriodDate := an.PeriodStartDate
	if linePeriodDate == "" {
		linePeriodDate = an.IssueDate
	}
	for i, line := range an.Lines {
		appendDocumentLine(root, "CreditNoteLine", "CreditedQuantity", i+1, line, an.CurrencyCode, an.DocumentTypeCode, domain.Identification{}, linePeriodDate)
	}

	return doc, nil
}

// appendNABillingReference adds cac:BillingReference to the original Support Document with
// schemeName "CUDS-SHA384" — unlike Credit/Debit Note, which reference an Invoice with
// "CUFE-SHA384".
func appendNABillingReference(parent *etree.Element, br domain.BillingReference) {
	ref := parent.CreateElement("cac:BillingReference").CreateElement("cac:InvoiceDocumentReference")
	ref.CreateElement("cbc:ID").SetText(br.Prefix + br.Number)
	uuid := ref.CreateElement("cbc:UUID")
	uuid.CreateAttr("schemeName", "CUDS-SHA384")
	uuid.SetText(br.CUFE)
	ref.CreateElement("cbc:IssueDate").SetText(br.IssueDate)
}
