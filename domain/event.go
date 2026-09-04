// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package domain

// EventParty is the reduced third-party shape ApplicationResponse events use for SenderParty/
// ReceiverParty (Technical Annex 1.9, section 6.5.4, groups AAF/AAG) — narrower than Party:
// just registration name, identification and tax scheme, no address/liability codes/contact.
type EventParty struct {
	Name           string
	Identification Identification
	TaxSchemeCode  string
	TaxSchemeName  string
}

// EventDocumentReference identifies the document (Invoice/CreditNote/DebitNote/SupportDocument)
// an event applies to (cac:DocumentResponse/cac:DocumentReference, sections 6.5.4/6.5.5).
type EventDocumentReference struct {
	Prefix           string
	Number           string
	CUFE             string // or CUDE/CUDS, depending on the referenced document's own type
	HashType         string // "CUFE-SHA384" / "CUDE-SHA384" / "CUDS-SHA384"
	DocumentTypeCode string // the referenced document's own type code, e.g. "01"
}

// EventReceiverPerson identifies the natural person who received the goods/service. Only used
// by the Acuse de Recibo event (section 6.5.5.3, group AAH11-18) — nil for every other type.
type EventReceiverPerson struct {
	Identification         Identification
	FirstName              string
	FamilyName             string
	JobTitle               string
	OrganizationDepartment string
}

// Event is the shared model for every DIAN "evento" (ApplicationResponse): Acuse de Recibo,
// Recibo del Bien y/o Servicio, Aceptación Expresa, Aceptación Tácita, and Reclamo (which
// embeds Event, see the Reclamo type). The ResponseCode/Description literals each event type
// requires are fixed by the annex, not caller data — builder.BuildAcuseRecibo/BuildReclamo/etc.
// set them; you don't.
//
// For all events except Aceptación Tácita ("034"), Sender is the receiver/acquirer of the
// referenced document and Receiver is its issuer — roles are inverted for Aceptación Tácita,
// where the issuer generates the event and DIAN itself is the recipient (section 6.5.5.7).
type Event struct {
	EnvironmentCode string

	ID        string // event's own consecutive number (AAD05)
	IssueDate string
	IssueTime string
	// Note is optional for most events; DIAN requires a specific templated statement for
	// Aceptación Tácita (034) — see event.TacitAcceptanceNote to build it correctly.
	Note string

	DocumentReference EventDocumentReference

	Sender   EventParty
	Receiver EventParty

	ReceiverPerson *EventReceiverPerson // Acuse de Recibo only; nil otherwise

	SoftwareProvider SoftwareProvider

	CUDE                 string
	SoftwareSecurityCode string
	// QRURL is built from the REFERENCED document's CUFE (qr.URL(EnvironmentCode,
	// DocumentReference.CUFE)), not from this event's own CUDE — section 6.5.4, AAB36.
	QRURL string
}

// Reclamo is the extra data the Reclamo event (ResponseCode "031") requires on top of Event:
// the rejection reason, from DIAN's catalog (section 13.3.11). cofacture doesn't validate
// catalog codes — same boundary as DiscrepancyResponse elsewhere in this module.
type Reclamo struct {
	Event
	RejectionListID string // cac:Response/cbc:ResponseCode/@listID
	RejectionName   string // cac:Response/cbc:ResponseCode/@name
}
