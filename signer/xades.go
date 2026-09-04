// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package signer

import (
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/beevik/etree"
)

// DIAN's signature policy. It is unique and fixed for all electronic documents — confirmed
// against two real invoices from different providers (2023 and 2026), same hash.
const (
	PolicyURL           = "https://facturaelectronica.dian.gov.co/politicadefirma/v2/politicadefirmav2.pdf"
	PolicyDescription   = "Política de firma para facturas electrónicas de la República de Colombia."
	PolicyHashSHA256B64 = "dMoMvtcG5aIzgYo0tIsSQeVJBDnUnfSOfBpxXrmor0Y="
)

// buildKeyInfo appends ds:KeyInfo (with the X.509 certificate in DER+base64) as a child of
// parent and returns it. parent must already be inserted in its final position within the
// document so canonicalization inherits the correct namespace context.
func (s *Signer) buildKeyInfo(parent *etree.Element, id string) *etree.Element {
	keyInfo := parent.CreateElement("ds:KeyInfo")
	keyInfo.CreateAttr("Id", id+"-keyinfo")
	cert := keyInfo.CreateElement("ds:X509Data").CreateElement("ds:X509Certificate")
	cert.SetText(base64.StdEncoding.EncodeToString(s.cert.Raw))
	return keyInfo
}

// buildSignedProperties appends ds:Object/xades:QualifyingProperties/xades:SignedProperties
// as a child of parent and returns the SignedProperties element. role is the value of
// xades:ClaimedRole ("supplier" for Invoice/CreditNote/DebitNote, "" for AttachedDocument).
func (s *Signer) buildSignedProperties(parent *etree.Element, id, role string, signingTime time.Time) *etree.Element {
	object := parent.CreateElement("ds:Object")
	qualifying := object.CreateElement("xades:QualifyingProperties")
	qualifying.CreateAttr("Target", "#"+id)

	signedProps := qualifying.CreateElement("xades:SignedProperties")
	signedProps.CreateAttr("Id", id+"-signedprops")

	ssp := signedProps.CreateElement("xades:SignedSignatureProperties")
	ssp.CreateElement("xades:SigningTime").SetText(signingTime.Format(time.RFC3339))

	certDigest := sha256.Sum256(s.cert.Raw)
	cert := ssp.CreateElement("xades:SigningCertificate").CreateElement("xades:Cert")
	digest := cert.CreateElement("xades:CertDigest")
	digest.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", digestMethodSHA256)
	digest.CreateElement("ds:DigestValue").SetText(base64.StdEncoding.EncodeToString(certDigest[:]))
	issuerSerial := cert.CreateElement("xades:IssuerSerial")
	issuerSerial.CreateElement("ds:X509IssuerName").SetText(s.cert.Issuer.String())
	issuerSerial.CreateElement("ds:X509SerialNumber").SetText(s.cert.SerialNumber.String())

	signaturePolicyID := ssp.CreateElement("xades:SignaturePolicyIdentifier").CreateElement("xades:SignaturePolicyId")
	sigPolicyID := signaturePolicyID.CreateElement("xades:SigPolicyId")
	sigPolicyID.CreateElement("xades:Identifier").SetText(PolicyURL)
	sigPolicyID.CreateElement("xades:Description").SetText(PolicyDescription)
	policyHash := signaturePolicyID.CreateElement("xades:SigPolicyHash")
	policyHash.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", digestMethodSHA256)
	policyHash.CreateElement("ds:DigestValue").SetText(PolicyHashSHA256B64)

	if role != "" {
		claimedRoles := ssp.CreateElement("xades:SignerRole").CreateElement("xades:ClaimedRoles")
		claimedRoles.CreateElement("xades:ClaimedRole").SetText(role)
	}

	return signedProps
}
