// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package cuds computes the Unique Support Document Code (Código Único de Documento Soporte,
// CUDS) per DIAN's Technical Annex 1.9, section 11.5.
//
// The CUDS formula differs from CUFE/CUDE's: instead of three fixed tax slots (VAT+"01",
// INC+"04", ICA+"03"), the CUDS uses a single CodImp+ValImp pair taken from the first element
// of HeaderTaxes. This behavior was verified against the official DIAN Support Document
// Toolkit v1.1 example (DocumentoSoporte-OperacionConResidente.xml, CUDS=c96a728f…23a) and
// matches the structure of the Support Document's QR content.
package cuds

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/diegofxm/cofacture/domain"
)

// Compute calculates the CUDS of a Support Document (InvoiceTypeCode "05").
//
// softwarePIN is the DIAN-authorized software PIN — the same value that appears in the QR
// content's PIN field and in sts:QRCode.
func Compute(doc domain.Invoice, softwarePIN string) string {
	var taxCode string
	var taxCents int64
	if len(doc.HeaderTaxes) > 0 {
		taxCode = doc.HeaderTaxes[0].TypeCode
		taxCents = doc.HeaderTaxes[0].TaxAmountCents
	}
	seed := doc.Prefix + doc.Number +
		doc.IssueDate + doc.IssueTime +
		domain.FormatCents(doc.Totals.LineExtensionCents) +
		taxCode + domain.FormatCents(taxCents) +
		domain.FormatCents(doc.Totals.PayableCents) +
		doc.Supplier.Identification.Number +
		doc.Customer.Identification.Number +
		softwarePIN +
		doc.EnvironmentCode
	sum := sha512.Sum384([]byte(seed))
	return hex.EncodeToString(sum[:])
}
