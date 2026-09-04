// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package securitycode computes sts:SoftwareSecurityCode per section 11.8 of Technical
// Annex 1.9. It is neither a CUFE nor a CUDE — it is the fingerprint of the DIAN-authorized
// software, applies equally to Invoice/CreditNote/DebitNote, and therefore has its own
// package instead of living inside cufe or cude.
package securitycode

import (
	"crypto/sha512"
	"encoding/hex"
)

// Compute calculates sts:SoftwareSecurityCode.
//
// softwareID and pin are sensitive values assigned by DIAN when the software is activated;
// they must never be logged or persisted in plain text outside of wherever they are already
// protected. documentID is the document's cbc:ID (Invoice/CreditNote/DebitNote/
// ApplicationResponse), i.e. Prefix+Number.
func Compute(softwareID, pin, documentID string) string {
	sum := sha512.Sum384([]byte(softwareID + pin + documentID))
	return hex.EncodeToString(sum[:])
}
