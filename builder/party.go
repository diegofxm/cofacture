// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strings"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
	ubl "github.com/diegofxm/cofacture/xml"
)

// appendAccountingParty adds AccountingSupplierParty or AccountingCustomerParty.
// showMerchantRegistration should only be true for the supplier (CorporateRegistrationScheme
// is the Chamber of Commerce registration, it doesn't apply to the customer). invoicePrefix is
// ignored when showMerchantRegistration is false.
//
// Verified against real invoices: cac:CorporateRegistrationScheme/cbc:ID is the invoice
// prefix (not something intrinsic to the third party), and cbc:Name is the mercantile
// registration number — omitted if the third party doesn't have one (e.g. a natural person
// with no mercantile registration).
func appendAccountingParty(parent *etree.Element, node string, p domain.Party, showMerchantRegistration bool, invoicePrefix string, showPhysicalLocation bool) {
	root := parent.CreateElement("cac:" + node)
	root.CreateElement("cbc:AdditionalAccountID").SetText(p.EntityTypeCode)

	party := root.CreateElement("cac:Party")

	// cbc:IndustryClassificationCode (CIIU) goes before PartyIdentification/PartyName in the
	// cac:Party sequence — confirmed against UBL-CommonAggregateComponents-2.1.xsd. Only the
	// supplier carries it (showMerchantRegistration is the same flag that distinguishes
	// supplier from customer in this function).
	if showMerchantRegistration && len(p.IndustryClassificationCodes) > 0 {
		party.CreateElement("cbc:IndustryClassificationCode").SetText(strings.Join(p.IndustryClassificationCodes, ";"))
	}

	if p.EntityTypeCode == "2" {
		partyID := party.CreateElement("cac:PartyIdentification").CreateElement("cbc:ID")
		setIdentificationAttrs(partyID, p.Identification)
		partyID.SetText(p.Identification.Number)
	}

	party.CreateElement("cac:PartyName").CreateElement("cbc:Name").SetText(p.Name)

	if showPhysicalLocation && p.Address.Line != "" {
		addr := party.CreateElement("cac:PhysicalLocation").CreateElement("cac:Address")
		appendAddressFields(addr, p.Address)
	}

	taxScheme := party.CreateElement("cac:PartyTaxScheme")
	taxScheme.CreateElement("cbc:RegistrationName").SetText(p.Name)
	companyID := taxScheme.CreateElement("cbc:CompanyID")
	setIdentificationAttrs(companyID, p.Identification)
	companyID.SetText(p.Identification.Number)
	if len(p.LiabilityCodes) > 0 {
		tlc := taxScheme.CreateElement("cbc:TaxLevelCode")
		tlc.CreateAttr("listName", p.TaxRegimeCode)
		tlc.SetText(strings.Join(p.LiabilityCodes, ";"))
	}
	if p.Address.Line != "" {
		regAddr := taxScheme.CreateElement("cac:RegistrationAddress")
		appendAddressFields(regAddr, p.Address)
	}
	scheme := taxScheme.CreateElement("cac:TaxScheme")
	scheme.CreateElement("cbc:ID").SetText(p.TaxSchemeCode)
	scheme.CreateElement("cbc:Name").SetText(p.TaxSchemeName)

	legal := party.CreateElement("cac:PartyLegalEntity")
	legal.CreateElement("cbc:RegistrationName").SetText(p.Name)
	legalID := legal.CreateElement("cbc:CompanyID")
	setIdentificationAttrs(legalID, p.Identification)
	legalID.SetText(p.Identification.Number)
	if showMerchantRegistration {
		crs := legal.CreateElement("cac:CorporateRegistrationScheme")
		crs.CreateElement("cbc:ID").SetText(invoicePrefix)
		if p.MerchantRegistrationNumber != nil {
			crs.CreateElement("cbc:Name").SetText(*p.MerchantRegistrationNumber)
		}
	}

	contact := party.CreateElement("cac:Contact")
	contact.CreateElement("cbc:Telephone").SetText(p.Phone)
	contact.CreateElement("cbc:ElectronicMail").SetText(p.Email)
}

func setIdentificationAttrs(el *etree.Element, id domain.Identification) {
	el.CreateAttr("schemeAgencyID", ubl.DianSchemeAgencyID)
	el.CreateAttr("schemeAgencyName", ubl.DianSchemeAgencyName)
	if id.TypeCode == ubl.IdentificationTypeNIT {
		el.CreateAttr("schemeID", id.VerificationCode)
	}
	el.CreateAttr("schemeName", id.TypeCode)
}

func appendAddressFields(addr *etree.Element, a domain.Address) {
	if a.CityCode != "" {
		addr.CreateElement("cbc:ID").SetText(a.CityCode)
		addr.CreateElement("cbc:CityName").SetText(a.CityName)
	}
	if a.PostalZone != "" {
		addr.CreateElement("cbc:PostalZone").SetText(a.PostalZone)
	}
	if a.StateCode != "" {
		addr.CreateElement("cbc:CountrySubentity").SetText(a.StateName)
		addr.CreateElement("cbc:CountrySubentityCode").SetText(a.StateCode)
	}
	addr.CreateElement("cac:AddressLine").CreateElement("cbc:Line").SetText(a.Line)
	if a.CountryCode != "" {
		country := addr.CreateElement("cac:Country")
		country.CreateElement("cbc:IdentificationCode").SetText(a.CountryCode)
		name := country.CreateElement("cbc:Name")
		name.CreateAttr("languageID", "es")
		name.SetText(a.CountryName)
	}
}
