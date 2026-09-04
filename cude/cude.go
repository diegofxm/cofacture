// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
//
// Package cude computes the Unique Electronic Document Code (Código Único de Documento
// Electrónico, CUDE) for Credit Notes and Debit Notes, per DIAN's Technical Annex 1.9
// (sections 11.4.3-11.4.6).
//
// The CUFE (Invoice) lives in package cufe, not here — the formulas are nearly identical, but
// they are distinct identifiers for distinct documents.
//
// This same function also computes the CUDE of a Documento Equivalente Electrónico (POS
// receipt, cinema ticket, toll receipt, etc. — InvoiceTypeCode 20/25/27/30/35/40/45/50/55/60)
// and of its own adjustment notes (types 93/94): the Documento Equivalente Electrónico
// Technical Annex V1.0 (Resolution 000165/2023, section 14.1.2/14.1.3) defines a CUDE with the
// exact same 14-field composition, order, and Software-PIN input as this one — DIAN reused the
// name and the formula, not just the name. There is no separate "dee" package for this on
// purpose: it would be a byte-for-byte duplicate of Compute below with a different name. If you
// are building a Documento Equivalente Electrónico or its adjustment note, this is the function
// you want — just pass the document's own Invoice-shaped fields the same way a CreditNote or
// DebitNote's embedded Invoice would.
package cude

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/diegofxm/cofacture/domain"
	"github.com/diegofxm/cofacture/internal/dianhash"
)

// Compute calculates the CUDE of a Credit Note, a Debit Note, a Documento Equivalente
// Electrónico, or a Documento Equivalente Electrónico adjustment note — see the package doc
// comment for why all four share this one function.
//
// The formula is identical to CUFE's in structure and field order (validated against the two
// official annex examples, one per note type) — the only real difference is that instead of
// the numbering range's technical key, the software's PIN is used ("Software-PIN", the same
// one used for securitycode.Compute). noteBase is the document's own Invoice-shaped data: the
// Invoice embedded in domain.CreditNote/domain.DebitNote (the promoted field, e.g.
// creditNote.Invoice) for a note, or the Documento Equivalente Electrónico's own domain.Invoice
// for that document type.
func Compute(noteBase domain.Invoice, softwarePIN string) string {
	sum := sha512.Sum384([]byte(dianhash.Seed(noteBase, softwarePIN)))
	return hex.EncodeToString(sum[:])
}
