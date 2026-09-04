// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
	ubl "github.com/diegofxm/cofacture/xml"
)

// BuildSupportDocument builds the XML tree of a Support Document for purchases made from
// non-obligated third parties (InvoiceTypeCode "05", CUDS-SHA384).
//
// The structure is Invoice (same root element and namespaces as BuildInvoice) with three key
// differences:
//  1. Inverted roles: Supplier = non-obligated third party (who sells to the issuer),
//     Customer = the issuing company (who acquires and generates the document).
//  2. WithholdingTaxTotal instead of (or in addition to) TaxTotal for withholdings.
//  3. InvoiceControl present (the Support Document has its own DIAN resolution, just like the
//     Invoice).
//
// The caller is responsible for computing the CUDS (package cuds), SoftwareSecurityCode and
// QRURL (package qr) before calling this function — inv's fields are serialized as-is.
func BuildSupportDocument(inv domain.Invoice) (*etree.Document, error) {
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

	// appendUBLExtensions handles InvoiceControl for DocumentTypeCode "05"
	appendUBLExtensions(root, inv)

	root.CreateElement("cbc:UBLVersionID").SetText("UBL 2.1")
	root.CreateElement("cbc:CustomizationID").SetText(inv.OperationTypeCode) // "10" or "11"
	root.CreateElement("cbc:ProfileID").SetText(inv.ProfileID)
	root.CreateElement("cbc:ProfileExecutionID").SetText(inv.EnvironmentCode)
	root.CreateElement("cbc:ID").SetText(inv.Prefix + inv.Number)

	uuid := root.CreateElement("cbc:UUID")
	uuid.CreateAttr("schemeID", inv.EnvironmentCode)
	uuid.CreateAttr("schemeName", inv.HashType) // "CUDS-SHA384"
	uuid.SetText(inv.CUFE)                      // the CUDS is stored in the model's CUFE field

	root.CreateElement("cbc:IssueDate").SetText(inv.IssueDate)
	root.CreateElement("cbc:IssueTime").SetText(inv.IssueTime)
	if inv.DueDate != "" {
		root.CreateElement("cbc:DueDate").SetText(inv.DueDate)
	}
	root.CreateElement("cbc:InvoiceTypeCode").SetText(inv.DocumentTypeCode) // "05"
	if inv.Note != "" {
		root.CreateElement("cbc:Note").SetText(inv.Note)
	}

	// Support Document: DocumentCurrencyCode with no list attributes (DSFC03 — the DS schema rejects them).
	root.CreateElement("cbc:DocumentCurrencyCode").SetText(inv.CurrencyCode)

	root.CreateElement("cbc:LineCountNumeric").SetText(strconv.Itoa(len(inv.Lines)))

	// Inverted roles: Supplier = non-obligated third party, Customer = the purchasing/issuing company.
	// The Support Document uses simpler party structures than the Invoice — dedicated functions per type.
	appendDSSupplierParty(root, inv.Supplier)
	appendDSCustomerParty(root, inv.Customer)

	for _, pm := range inv.PaymentMeans {
		appendPaymentMean(root, pm)
	}

	appendTaxTotal(root, inv.HeaderTaxes, inv.CurrencyCode)
	appendWithholdingTaxTotal(root, inv.WithholdingTaxes, inv.CurrencyCode)
	appendMonetaryTotal(root, "LegalMonetaryTotal", inv.Totals, inv.CurrencyCode, inv.DocumentTypeCode)

	// Per-line period date: the document's PeriodStartDate if specified, otherwise IssueDate.
	// cac:InvoicePeriod is required on every InvoiceLine of the Support Document (DSFC01).
	linePeriodDate := inv.PeriodStartDate
	if linePeriodDate == "" {
		linePeriodDate = inv.IssueDate
	}
	for i, line := range inv.Lines {
		appendDocumentLine(root, "InvoiceLine", "InvoicedQuantity", i+1, line, inv.CurrencyCode, inv.DocumentTypeCode, domain.Identification{}, linePeriodDate)
	}

	return doc, nil
}

// appendDSSupplierParty generates the Support Document's AccountingSupplierParty.
// DIAN requires schemeName="31" (NIT) for the non-obligated third party's CompanyID in all
// cases — including natural persons — based on behavior verified against a real, DIAN-accepted
// Support Document (DS-real.xml). TaxLevelCode has no listName attribute (also verified against
// the real document).
func appendDSSupplierParty(root *etree.Element, p domain.Party) {
	ap := root.CreateElement("cac:AccountingSupplierParty")
	ap.CreateElement("cbc:AdditionalAccountID").SetText(p.EntityTypeCode)
	party := ap.CreateElement("cac:Party")

	party.CreateElement("cac:PartyName").CreateElement("cbc:Name").SetText(p.Name)

	if p.Address.Line != "" {
		addr := party.CreateElement("cac:PhysicalLocation").CreateElement("cac:Address")
		appendAddressFields(addr, p.Address)
	}

	taxScheme := party.CreateElement("cac:PartyTaxScheme")
	taxScheme.CreateElement("cbc:RegistrationName").SetText(p.Name)
	companyID := taxScheme.CreateElement("cbc:CompanyID")
	companyID.CreateAttr("schemeAgencyID", ubl.DianSchemeAgencyID)
	companyID.CreateAttr("schemeAgencyName", ubl.DianSchemeAgencyName)
	// DIAN requires schemeName="31" (NIT) for the non-obligated third party; schemeID is the check digit.
	if p.Identification.VerificationCode != "" {
		companyID.CreateAttr("schemeID", p.Identification.VerificationCode)
	}
	companyID.CreateAttr("schemeName", "31")
	companyID.SetText(p.Identification.Number)
	if len(p.LiabilityCodes) > 0 {
		taxScheme.CreateElement("cbc:TaxLevelCode").SetText(strings.Join(p.LiabilityCodes, ";"))
	}
	scheme := taxScheme.CreateElement("cac:TaxScheme")
	scheme.CreateElement("cbc:ID").SetText(p.TaxSchemeCode)
	scheme.CreateElement("cbc:Name").SetText(p.TaxSchemeName)
}

// appendDSCustomerParty generates the Support Document's AccountingCustomerParty.
// The purchasing/issuing company carries ONLY PartyTaxScheme — no PartyIdentification,
// PhysicalLocation, RegistrationAddress, PartyLegalEntity or Contact.
// TaxLevelCode carries no listName attribute.
// Verified against DIAN's Support Document Toolkit v1.1.
func appendDSCustomerParty(root *etree.Element, p domain.Party) {
	ap := root.CreateElement("cac:AccountingCustomerParty")
	ap.CreateElement("cbc:AdditionalAccountID").SetText(p.EntityTypeCode)
	party := ap.CreateElement("cac:Party")

	party.CreateElement("cac:PartyName").CreateElement("cbc:Name").SetText(p.Name)

	if p.Address.Line != "" {
		addr := party.CreateElement("cac:PhysicalLocation").CreateElement("cac:Address")
		appendAddressFields(addr, p.Address)
	}

	taxScheme := party.CreateElement("cac:PartyTaxScheme")
	taxScheme.CreateElement("cbc:RegistrationName").SetText(p.Name)
	companyID := taxScheme.CreateElement("cbc:CompanyID")
	setIdentificationAttrs(companyID, p.Identification)
	companyID.SetText(p.Identification.Number)
	if len(p.LiabilityCodes) > 0 {
		taxScheme.CreateElement("cbc:TaxLevelCode").SetText(strings.Join(p.LiabilityCodes, ";"))
	}
	scheme := taxScheme.CreateElement("cac:TaxScheme")
	scheme.CreateElement("cbc:ID").SetText(p.TaxSchemeCode)
	scheme.CreateElement("cbc:Name").SetText(p.TaxSchemeName)
}
