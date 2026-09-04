// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package domain

// AttachedPartyInfo is the reduced version of a third party used by AttachedDocument
// (SenderParty/ReceiverParty) — it only carries basic tax data, unlike Party (which also has
// address, contact and legal representation, used in AccountingSupplierParty/
// AccountingCustomerParty within the wrapped document).
type AttachedPartyInfo struct {
	Name           string
	Identification Identification
	TaxRegimeCode  string // TaxLevelCode's listName
	LiabilityCodes []string
	TaxSchemeCode  string
	TaxSchemeName  string
}

// ValidationResult is DIAN's validation response for the wrapped document, embedded in
// cac:ParentDocumentLineReference. The technical annex requires at least one occurrence (AE38,
// "1..N") — the AttachedDocument is an artifact produced after validation, not something built
// before it.
type ValidationResult struct {
	LineID string // sequence number, normally "1"

	DocumentID        string // ID of the referenced document (Prefix+Number)
	DocumentCUFE      string // or CUDE
	DocumentHashType  string // "CUFE-SHA384" or "CUDE-SHA384"
	DocumentIssueDate string

	// ApplicationResponseXML is the full ApplicationResponse returned by DIAN, verbatim, for
	// the CDATA of cac:Attachment/cac:ExternalReference/cbc:Description.
	ApplicationResponseXML string

	ValidatorID          string // fixed: "Unidad Especial Dirección de Impuestos y Aduanas Nacionales"
	ValidationResultCode string // e.g. "02"
	ValidationDate       string
	ValidationTime       string
}

// AttachedDocument is the electronic container delivered to the acquirer: it wraps the signed
// document (Invoice/CreditNote/DebitNote) together with DIAN's validation response.
type AttachedDocument struct {
	EnvironmentCode string

	// ID is the "generator's own consecutive number" (AE04b) — it is NOT the wrapped
	// document's CUFE; the two are different values even though some providers conflate them
	// in practice.
	ID string

	IssueDate string // container generation date (>= the wrapped document's IssueDate)
	IssueTime string

	// ParentDocumentID is the cbc:ID of the wrapped document (Prefix+Number), not of the
	// container itself.
	ParentDocumentID string

	Sender   AttachedPartyInfo
	Receiver AttachedPartyInfo

	// AttachmentXML is the signed XML of the wrapped document (Invoice/CreditNote/
	// DebitNote), verbatim, for the CDATA of cac:Attachment/cac:ExternalReference/
	// cbc:Description.
	AttachmentXML string

	ValidationResults []ValidationResult
}
