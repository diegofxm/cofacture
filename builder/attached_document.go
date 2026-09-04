// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"strings"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
	ubl "github.com/diegofxm/cofacture/xml"
)

const (
	attachedDocumentNamespace       = "urn:oasis:names:specification:ubl:schema:xsd:AttachedDocument-2"
	attachedDocumentCustomizationID = "Documentos adjuntos"
	invoiceContainerProfileID       = "Factura Electrónica de Venta"
	invoiceContainerDocumentType    = "Contenedor de Factura Electrónica"
	creditNoteContainerProfileID    = "Nota de Crédito de Venta"
	creditNoteContainerDocumentType = "Contenedor de Nota de Crédito"
	debitNoteContainerProfileID     = "Nota de Débito de Venta"
	debitNoteContainerDocumentType  = "Contenedor de Nota de Débito"
)

// BuildInvoiceAttachedDocument builds the AttachedDocument (electronic container) that wraps
// an already-signed Electronic Sales Invoice, together with DIAN's validation response.
//
// It only makes sense to call this after receiving that response (GetStatusZip): the technical
// annex requires at least one cac:ParentDocumentLineReference (AE38, "1..N"), and the two real
// invoices used to verify this structure both confirm the container is always delivered with
// the validation already embedded, never before — the AttachedDocument is not what gets sent
// to DIAN (that's the signed Invoice, inside a ZIP, via SendBillSync/SendBillAsync); it's what
// gets delivered to the acquirer afterward.
func BuildInvoiceAttachedDocument(ad domain.AttachedDocument) (*etree.Document, error) {
	return buildAttachedDocument(ad, invoiceContainerProfileID, invoiceContainerDocumentType)
}

// BuildCreditNoteAttachedDocument builds the AttachedDocument for an already-signed and
// DIAN-accepted Credit Note. Same contract as BuildInvoiceAttachedDocument.
func BuildCreditNoteAttachedDocument(ad domain.AttachedDocument) (*etree.Document, error) {
	return buildAttachedDocument(ad, creditNoteContainerProfileID, creditNoteContainerDocumentType)
}

// BuildDebitNoteAttachedDocument builds the AttachedDocument for an already-signed and
// DIAN-accepted Debit Note. Same contract as BuildInvoiceAttachedDocument.
func BuildDebitNoteAttachedDocument(ad domain.AttachedDocument) (*etree.Document, error) {
	return buildAttachedDocument(ad, debitNoteContainerProfileID, debitNoteContainerDocumentType)
}

func buildAttachedDocument(ad domain.AttachedDocument, profileID, docType string) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8" standalone="no"`)

	root := doc.CreateElement("AttachedDocument")
	root.CreateAttr("xmlns", attachedDocumentNamespace)
	root.CreateAttr("xmlns:cac", ubl.NSCac)
	root.CreateAttr("xmlns:cbc", ubl.NSCbc)
	root.CreateAttr("xmlns:ds", ubl.NSDs)
	root.CreateAttr("xmlns:ext", ubl.NSExt)
	root.CreateAttr("xmlns:xades", ubl.NSXades)
	root.CreateAttr("xmlns:xades141", ubl.NSXades141)

	// Placeholder for the AttachedDocument's own signature (signer.Sign with role="").
	root.CreateElement("ext:UBLExtensions").CreateElement("ext:UBLExtension").CreateElement("ext:ExtensionContent")

	root.CreateElement("cbc:UBLVersionID").SetText("UBL 2.1")
	root.CreateElement("cbc:CustomizationID").SetText(attachedDocumentCustomizationID)
	root.CreateElement("cbc:ProfileID").SetText(profileID)
	root.CreateElement("cbc:ProfileExecutionID").SetText(ad.EnvironmentCode)
	root.CreateElement("cbc:ID").SetText(ad.ID)
	root.CreateElement("cbc:IssueDate").SetText(ad.IssueDate)
	root.CreateElement("cbc:IssueTime").SetText(ad.IssueTime)
	root.CreateElement("cbc:DocumentType").SetText(docType)
	root.CreateElement("cbc:ParentDocumentID").SetText(ad.ParentDocumentID)

	appendAttachedParty(root, "SenderParty", ad.Sender)
	appendAttachedParty(root, "ReceiverParty", ad.Receiver)

	attachment := root.CreateElement("cac:Attachment").CreateElement("cac:ExternalReference")
	attachment.CreateElement("cbc:MimeCode").SetText("text/xml")
	attachment.CreateElement("cbc:EncodingCode").SetText("UTF-8")
	attachment.CreateElement("cbc:Description").CreateCData(ad.AttachmentXML)

	for _, vr := range ad.ValidationResults {
		appendParentDocumentLineReference(root, vr)
	}

	return doc, nil
}

func appendAttachedParty(parent *etree.Element, node string, p domain.AttachedPartyInfo) {
	taxScheme := parent.CreateElement("cac:" + node).CreateElement("cac:PartyTaxScheme")
	taxScheme.CreateElement("cbc:RegistrationName").SetText(p.Name)
	companyID := taxScheme.CreateElement("cbc:CompanyID")
	setIdentificationAttrs(companyID, p.Identification)
	companyID.SetText(p.Identification.Number)
	if len(p.LiabilityCodes) > 0 {
		tlc := taxScheme.CreateElement("cbc:TaxLevelCode")
		tlc.CreateAttr("listName", p.TaxRegimeCode)
		tlc.SetText(strings.Join(p.LiabilityCodes, ";"))
	}
	scheme := taxScheme.CreateElement("cac:TaxScheme")
	scheme.CreateElement("cbc:ID").SetText(p.TaxSchemeCode)
	scheme.CreateElement("cbc:Name").SetText(p.TaxSchemeName)
}

func appendParentDocumentLineReference(parent *etree.Element, vr domain.ValidationResult) {
	pdlr := parent.CreateElement("cac:ParentDocumentLineReference")
	pdlr.CreateElement("cbc:LineID").SetText(vr.LineID)

	docRef := pdlr.CreateElement("cac:DocumentReference")
	docRef.CreateElement("cbc:ID").SetText(vr.DocumentID)
	uuid := docRef.CreateElement("cbc:UUID")
	uuid.CreateAttr("schemeName", vr.DocumentHashType)
	uuid.SetText(vr.DocumentCUFE)
	docRef.CreateElement("cbc:IssueDate").SetText(vr.DocumentIssueDate)
	docRef.CreateElement("cbc:DocumentType").SetText("ApplicationResponse")

	attachment := docRef.CreateElement("cac:Attachment").CreateElement("cac:ExternalReference")
	attachment.CreateElement("cbc:MimeCode").SetText("text/xml")
	attachment.CreateElement("cbc:EncodingCode").SetText("UTF-8")
	attachment.CreateElement("cbc:Description").CreateCData(vr.ApplicationResponseXML)

	result := docRef.CreateElement("cac:ResultOfVerification")
	result.CreateElement("cbc:ValidatorID").SetText(vr.ValidatorID)
	result.CreateElement("cbc:ValidationResultCode").SetText(vr.ValidationResultCode)
	result.CreateElement("cbc:ValidationDate").SetText(vr.ValidationDate)
	result.CreateElement("cbc:ValidationTime").SetText(vr.ValidationTime)
}
