// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package event computes the Unique Electronic Document Code (CUDE) for ApplicationResponse
// events per DIAN's Technical Annex 1.9, section 11.5. This is a different formula from the
// CUDE the cude package computes for Credit/Debit Notes — DIAN reuses the same name ("CUDE")
// for both, but the field composition is not the same, so they live in separate packages.
package event

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

// DIAN's response-code catalog for events (Technical Annex 1.9, section 13.3.1) — the five
// values cofacture's builder package knows how to produce. Not a general-purpose catalog: it's
// exactly the five constants builder.BuildAcuseRecibo/BuildReclamo/BuildRecibidoBien/
// BuildAceptacionExpresa/BuildAceptacionTacita each hardcode internally.
const (
	ResponseCodeAcuseRecibo       = "030"
	ResponseCodeReclamo           = "031"
	ResponseCodeRecibidoBien      = "032"
	ResponseCodeAceptacionExpresa = "033"
	ResponseCodeAceptacionTacita  = "034"
)

// Compute calculates the CUDE of an ApplicationResponse event.
//
// senderNIT is the event generator's identification (Sender); receiverNIT is the recipient's
// (Receiver). documentID and documentTypeCode belong to the REFERENCED document (Prefix+Number
// and its own type code, e.g. "01"), not to the event itself. softwarePIN is the DIAN-assigned
// software PIN, never part of the XML.
//
// Formula (section 11.5): SHA-384(Num_DE+Fec_Emi+Hor_Emi+NitFE+DocAdq+ResponseCode+ID+
// DocumentTypeCode+Software-PIN). Verified against the annex's own worked example (11.5.1):
// numDE="1", issueDate="2019-04-30", issueTime="19:48:50-05:00", senderNIT="99998888",
// receiverNIT="800197268", responseCode="030", documentID="FE123", documentTypeCode="01",
// softwarePIN="11111" → CUDE
// 0d91ba25b01f5e7dbda870a11b274501d3a62a73e91932c473c86c93f12a142a2ac45876efcde3e679024a01c0be41f9
func Compute(numDE, issueDate, issueTime, senderNIT, receiverNIT, responseCode, documentID, documentTypeCode, softwarePIN string) string {
	seed := numDE + issueDate + issueTime + senderNIT + receiverNIT + responseCode + documentID + documentTypeCode + softwarePIN
	sum := sha512.Sum384([]byte(seed))
	return hex.EncodeToString(sum[:])
}

// TacitAcceptanceNote builds the sworn-statement text DIAN requires in cbc:Note when
// registering Aceptación Tácita (event "034"), section 6.5.5.7 — the "sin mandatario" template
// (a natural or legal person sending the event directly, not through a proxy/mandatario). The
// annex also defines a second "con mandatario" template for when a proxy sends the event on
// the issuer's behalf; that variant isn't covered here yet.
//
// recibidoBienID/recibidoBienCUDE identify the earlier Recibo del Bien y/o Servicio event
// (ResponseCode "032") this statement refers to — DIAN requires that event to already exist
// before Aceptación Tácita can be registered. acquirerName/acquirerNIT identify the acquirer
// who neither accepted, rejected, nor claimed against the invoice within the 3 business days.
func TacitAcceptanceNote(recibidoBienID, recibidoBienCUDE, acquirerName, acquirerNIT string) string {
	return fmt.Sprintf(
		"Manifiesto bajo la gravedad de juramento que transcurridos 3 días hábiles contados desde la creación del Recibo de bienes y servicios %s con CUDE %s, el adquirente %s identificado con NIT %s no manifestó expresamente la aceptación o rechazo de la referida factura, ni reclamó en contra de su contenido.",
		recibidoBienID, recibidoBienCUDE, acquirerName, acquirerNIT,
	)
}
