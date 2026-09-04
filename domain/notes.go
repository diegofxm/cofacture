// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

// BillingReference identifies the invoice that a Credit/Debit Note corrects
// (cac:BillingReference) — always required in both, unlike DiscrepancyResponse.
type BillingReference struct {
	Prefix    string
	Number    string
	CUFE      string // hash of the referenced document (not this note's own CUFE/CUDE)
	IssueDate string

	// HashType is the referenced document's own hash scheme — "CUFE-SHA384" when a Credit/Debit
	// Note corrects a regular Invoice, "CUDE-SHA384" when it corrects a Documento Equivalente
	// Electrónico (e.g. a POS ticket, DocumentTypeCode "93"/"94"). It must match the referenced
	// document's actual HashType, not this note's own — the two can differ.
	HashType string
}

// DiscrepancyResponse is the note's reason (cac:DiscrepancyResponse) — required when the note
// references a specific invoice (operation "20" for Credit Note, "30" for Debit Note,
// operation type catalog). ResponseCode uses the credit/debit note concept catalog.
type DiscrepancyResponse struct {
	ReferenceID  string // normally Prefix+Number of the referenced invoice
	ResponseCode string
	Description  string
}

// CreditNote is a Credit Note. It shares almost every field with Invoice — UBL treats them as
// distinct documents, but what actually changes is the document type, the reference to the
// corrected invoice, and the reason. That's why Invoice is reused instead of repeating ~20
// fields.
//
// The inherited CUFE field actually holds this note's CUDE (the same slot DIAN uses —
// cbc:UUID — for any of the electronic documents; HashType is what distinguishes
// "CUFE-SHA384" from "CUDE-SHA384", not the field name in this model).
type CreditNote struct {
	Invoice

	CreditNoteTypeCode  string // Credit Note type catalog (cbc:CreditNoteTypeCode)
	BillingReference    BillingReference
	DiscrepancyResponse *DiscrepancyResponse // nil if the operation does not require a reason
}

// DebitNote is a Debit Note. Structurally identical to CreditNote — the real differences from
// the builder's perspective are that it uses cac:RequestedMonetaryTotal instead of
// cac:LegalMonetaryTotal, and the technical annex defines no equivalent to
// cbc:CreditNoteTypeCode for Debit Note (verified: no such element appears between
// DocumentCurrencyCode and LineCountNumeric in section 8.4 of the annex).
//
// DIAN business rule confirmed against a real submission (2026-09-02, rule DAJ48 "Debe ser
// 31"): the issuer's own Supplier.Identification.TypeCode must be "31" (NIT) — even for a
// natural person whose Invoice/CreditNote both accept "13" (cédula) with zero complaints. A
// related rule (DAJ39) also requires Supplier.TaxSchemeCode/"Name to be a real regime (e.g.
// "01"/"IVA") once identified via NIT — "ZZ"/"No aplica" (otherwise fine on Invoice/CreditNote)
// is rejected here. This package does not enforce either (same caller-trust boundary as
// everything else here); the caller is responsible for setting both correctly.
type DebitNote struct {
	Invoice

	BillingReference    BillingReference
	DiscrepancyResponse *DiscrepancyResponse
}

// AdjustmentNote is the Adjustment Note to the Support Document (InvoiceTypeCode "95").
// It is to the Support Document (type 05) what Credit/Debit Notes are to the Invoice: it
// allows correcting or voiding a previously issued Support Document. The roles are the same
// as in the Support Document: Supplier = non-obligated third party (SNO), Customer = the
// purchasing/issuing company (ABS). Uses CUDS-SHA384 (same formula as the Support Document).
// The inherited CUFE field holds this note's CUDS (schemeName "CUDS-SHA384").
type AdjustmentNote struct {
	Invoice

	BillingReference    BillingReference // Reference to the original Support Document (UUID with schemeName CUDS-SHA384)
	DiscrepancyResponse *DiscrepancyResponse
}
