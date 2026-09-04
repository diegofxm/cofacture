// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package cufe computes the Unique Electronic Invoice Code (Código Único de Factura
// Electrónica, CUFE) per DIAN's Technical Annex 1.9 (Resolution 000165/2023, section 11.2).
//
// The CUDE (Credit/Debit Notes) lives in package cude, not here — the formulas are nearly
// identical, but they are distinct identifiers for distinct documents; keeping them in
// separate packages avoids someone searching for "cude" and only finding "cufe".
package cufe

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/diegofxm/cofacture/domain"
	"github.com/diegofxm/cofacture/internal/dianhash"
)

// Compute calculates the CUFE of an Electronic Sales Invoice.
//
// technicalKey is the authorized numbering range's "technical key" (ClTec) — it is obtained
// from DIAN's GetNumberingRange web service and never travels inside the XML, so it is not
// part of domain.Invoice/domain.NumberingRange. Whoever calls Compute is responsible for
// obtaining and storing it securely.
//
// Formula (section 11.2): SHA-384(NumFac+FecFac+HorFac+ValFac+CodImp1+ValImp1+CodImp2+
// ValImp2+CodImp3+ValImp3+ValTot+NitOFE+NumAdq+ClTec+TipoAmbiente), where CodImp1/2/3 are
// fixed "01"/"04"/"03" (VAT/INC/ICA) and ValImpN is "0.00" when that tax does not apply.
func Compute(inv domain.Invoice, technicalKey string) string {
	sum := sha512.Sum384([]byte(dianhash.Seed(inv, technicalKey)))
	return hex.EncodeToString(sum[:])
}
