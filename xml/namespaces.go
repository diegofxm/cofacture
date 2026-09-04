// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
//
// Package xml holds the UBL 2.1 + DIAN extension namespace constants shared by all builders.
package xml

const (
	NSInvoiceDefault      = "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
	NSCreditNote          = "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2"
	NSDebitNote           = "urn:oasis:names:specification:ubl:schema:xsd:DebitNote-2"
	NSApplicationResponse = "urn:oasis:names:specification:ubl:schema:xsd:ApplicationResponse-2"
	NSCac                 = "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
	NSCbc                 = "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
	NSDs                  = "http://www.w3.org/2000/09/xmldsig#"
	NSExt                 = "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2"
	NSSts                 = "dian:gov:co:facturaelectronica:Structures-2-1"
	NSXades               = "http://uri.etsi.org/01903/v1.3.2#"
	NSXades141            = "http://uri.etsi.org/01903/v1.4.1#"
	NSXsi                 = "http://www.w3.org/2001/XMLSchema-instance"

	SchemaLocationInvoice = NSInvoiceDefault + " http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/UBL-Invoice-2.1.xsd"
	// SchemaLocationApplicationResponse follows the standard UBL 2.1 maindoc namespace pattern
	// (same shape as SchemaLocationInvoice) — unlike the rest of this file, it has not been
	// cross-checked against a real, DIAN-accepted ApplicationResponse event document, only
	// against the Technical Annex's field tables (section 6.5.4) and the OASIS UBL 2.1 schema.
	SchemaLocationApplicationResponse = NSApplicationResponse + " http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/UBL-ApplicationResponse-2.1.xsd"

	// DianAuthorizationProviderID is DIAN's fixed NIT as the authorization provider.
	DianAuthorizationProviderID = "800197268"
	// DianSchemeAgencyID identifies DIAN as the schema (catalog) agency.
	DianSchemeAgencyID   = "195"
	DianSchemeAgencyName = "CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)"

	// IdentificationTypeNIT is the catalog code that triggers the check digit.
	IdentificationTypeNIT = "31"
)
