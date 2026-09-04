// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package qr builds the QR code URL required in the graphic representation of DIAN electronic
// documents (Technical Annex 1.9, section 11.7.1).
package qr

import (
	"fmt"
	"strconv"

	"github.com/diegofxm/cofacture/domain"
)

const (
	habilitacionBaseURL = "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr"
	produccionBaseURL   = "https://catalogo-vpfe.dian.gov.co/document/searchqr"
)

// URL builds the QR URL from the CUFE (or CUDE) and the environment code ("1" production, "2"
// certification/testing — the same value as cbc:ProfileExecutionID). DIAN uses different
// domains per environment; confirmed against two real production invoices and the technical
// annex text.
func URL(environmentCode, documentKey string) string {
	base := produccionBaseURL
	if environmentCode == "2" {
		base = habilitacionBaseURL
	}
	return base + "?documentkey=" + documentKey
}

// SupportDocumentURL builds the Support Document's QR URL.
// Uses the same searchqr endpoint as Invoice/CreditNote/DebitNote (verified: FindDocument does
// not redirect).
func SupportDocumentURL(environmentCode, cuds string) string {
	return URL(environmentCode, cuds)
}

// AdjustmentNoteContent builds the full QR content for the Adjustment Note to the Support
// Document (InvoiceTypeCode "95"). Follows the same pattern as SupportDocumentContent (a
// multi-line text block followed by the URL), adapted for the Adjustment Note document type.
func AdjustmentNoteContent(inv domain.Document, cuds, softwarePIN string) string {
	var codImp, valImp string
	for _, t := range inv.HeaderTaxes {
		if t.TypeCode == "01" {
			codImp = t.TypeCode
			valImp = domain.FormatCents(t.TaxAmountCents)
			break
		}
	}
	if codImp == "" && len(inv.HeaderTaxes) > 0 {
		codImp = inv.HeaderTaxes[0].TypeCode
		valImp = domain.FormatCents(inv.HeaderTaxes[0].TaxAmountCents)
	}
	if codImp == "" {
		codImp = "01"
		valImp = "0.00"
	}

	url := URL(inv.EnvironmentCode, cuds)

	ambLabel := strconv.Itoa(func() int {
		if inv.EnvironmentCode == "2" {
			return 2
		}
		return 1
	}())

	return fmt.Sprintf(
		"N°NotaAjuste=%s\nFecha=%s\nHora=%s\nValNA=%s\nCodImp=%s\nValImp=%s\nValTot=%s\nNumSNO=%s\nNITABS=%s\nPIN:%s\nAmb:%s\nCUDS=%s\nURL=%s",
		inv.Prefix+inv.Number,
		inv.IssueDate,
		inv.IssueTime,
		domain.FormatCents(inv.Totals.LineExtensionCents),
		codImp,
		valImp,
		domain.FormatCents(inv.Totals.PayableCents),
		inv.Supplier.Identification.Number,
		inv.Customer.Identification.Number,
		softwarePIN,
		ambLabel,
		cuds,
		url,
	)
}

// SupportDocumentContent builds the full QR content for the Support Document (InvoiceTypeCode
// "05"). Unlike Invoice/CreditNote/DebitNote, whose QR is just a URL, the Support Document
// requires a multi-line text block followed by the URL (Technical Annex 1.9, section 11.7.1).
//
// softwarePIN is the authorized software PIN — the same one used to compute the CUDS.
// cuds is the already-computed CUDS (hex SHA-384).
//
// Roles in the Support Document: Supplier = non-obligated third party (NumSNO), Customer =
// issuing company (NITABS).
func SupportDocumentContent(inv domain.Document, cuds, softwarePIN string) string {
	var codImp, valImp string
	for _, t := range inv.HeaderTaxes {
		if t.TypeCode == "01" { // VAT
			codImp = t.TypeCode
			valImp = domain.FormatCents(t.TaxAmountCents)
			break
		}
	}
	if codImp == "" && len(inv.HeaderTaxes) > 0 {
		codImp = inv.HeaderTaxes[0].TypeCode
		valImp = domain.FormatCents(inv.HeaderTaxes[0].TaxAmountCents)
	}
	if codImp == "" {
		codImp = "01"
		valImp = "0.00"
	}

	url := URL(inv.EnvironmentCode, cuds)

	ambLabel := strconv.Itoa(func() int {
		if inv.EnvironmentCode == "2" {
			return 2
		}
		return 1
	}())

	return fmt.Sprintf(
		"N°DocSoporte=%s\nFecha=%s\nHora=%s\nValDS=%s\nCodImp=%s\nValImp=%s\nValTot=%s\nNumSNO=%s\nNITABS=%s\nPIN:%s\nAmb:%s\nCUDS=%s\nURL=%s",
		inv.Prefix+inv.Number,
		inv.IssueDate,
		inv.IssueTime,
		domain.FormatCents(inv.Totals.LineExtensionCents),
		codImp,
		valImp,
		domain.FormatCents(inv.Totals.PayableCents),
		inv.Supplier.Identification.Number,
		inv.Customer.Identification.Number,
		softwarePIN,
		ambLabel,
		cuds,
		url,
	)
}
