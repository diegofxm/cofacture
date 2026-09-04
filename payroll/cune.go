// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package payroll

import (
	"crypto/sha512"
	"fmt"

	"github.com/diegofxm/cofacture/domain"
)

// Cune computes the Unique Electronic Payroll Code (Código Único de Nómina Electrónica, CUNE)
// per section 8.1 of DIAN's Technical Annex:
//
//	SHA-384(NumNE + FecNE + HorNE + ValDev + ValDed + ValTolNE + NitNE + DocEmp + TipoXML + SoftwarePIN + TipAmb)
//
// valDevCents, valDedCents and valTolNECents are formatted with domain.FormatCents — the same
// function Build uses for DevengadosTotal/DeduccionesTotal/ComprobanteTotal — so the hashed
// amounts and the amounts written to the XML can never drift apart. The result is the
// 96-character hex SHA-384 that goes into InformacionGeneral/@CUNE.
func Cune(numNE, fecNE, horNE string, valDevCents, valDedCents, valTolNECents int64, nitNE, docEmp, tipoXML, softwarePIN, tipAmb string) string {
	input := numNE + fecNE + horNE +
		domain.FormatCents(valDevCents) + domain.FormatCents(valDedCents) + domain.FormatCents(valTolNECents) +
		nitNE + docEmp + tipoXML + softwarePIN + tipAmb
	sum := sha512.Sum384([]byte(input))
	return fmt.Sprintf("%x", sum)
}
