// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package zip

import "fmt"

// DocumentKind identifies the document type for the XML file name
// (Technical Annex 1.9, section 6.5.7).
type DocumentKind string

const (
	KindInvoice             DocumentKind = "fv" // Sales Invoice
	KindCreditNote          DocumentKind = "nc" // Credit Note
	KindDebitNote           DocumentKind = "nd" // Debit Note
	KindSupportDocument     DocumentKind = "ds" // Support Document
	KindAdjustmentNote      DocumentKind = "na" // Adjustment Note to the Support Document
	KindApplicationResponse DocumentKind = "ar"
	KindAttachedDocument    DocumentKind = "ad"
)

// SoftwarePropioCode is the Technology Provider code ("ppp") for issuers using their own
// software instead of operating as a Technology Provider for third parties (section 6.5.8,
// note).
const SoftwarePropioCode = "000"

// DocumentFileName builds the XML file name required by section 6.5.7:
//
//	{kind}{NIT without check digit, 10 digits}{PT code, 3 digits}{year, 2 digits}{consecutive, 8 hex}.xml
//
// nit must be passed without the check digit. ptCode is the 3-digit Technology Provider code
// assigned by DIAN (SoftwarePropioCode for own software). year is the last 2 digits of the
// calendar year. consecutive is the running count of files sent for the corresponding type —
// the annex requires resetting it to 1 every January 1st; keeping track of it is the
// responsibility of whoever orchestrates the sending (persistent state), not this package,
// which only formats.
//
// Note: the text of section 6.5.7/6.5.8 describes the consecutive as hexadecimal ("in the
// range 00000001 <= FFFFFFFF"), but the illustrative example the annex itself gives for the
// "eleventh invoice" shows "00000011" — which in hexadecimal would be the 17th, not the 11th.
// This is an inconsistency in the document, not something that can be resolved from here. This
// function follows the rule's text (hex) because that is the normative reading; if DIAN's
// certification environment turns out to expect decimal, this will need adjusting.
func DocumentFileName(kind DocumentKind, nit, ptCode string, year int, consecutive uint32) string {
	return fmt.Sprintf("%s%010s%s%02d%08X.xml", kind, nit, ptCode, year%100, consecutive)
}

// PackageFileName builds the ZIP file name required by section 6.5.8:
//
//	z{NIT without check digit, 10 digits}{PT code, 3 digits}{year, 2 digits}{consecutive, 8 hex}.zip
//
// The consecutive here is the compressed package's own, distinct from the individual XML
// files' consecutive in DocumentFileName.
func PackageFileName(nit, ptCode string, year int, consecutive uint32) string {
	return fmt.Sprintf("z%010s%s%02d%08X.zip", nit, ptCode, year%100, consecutive)
}
