// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package signer signs UBL documents with XAdES-EPES, as required by DIAN.
//
// The structure (three ds:Reference elements — document, KeyInfo and SignedProperties — plus
// DIAN's fixed signature policy) was verified byte-for-byte against two real, accepted
// electronic invoices from two different technology providers, not inferred solely from the
// technical annex.
package signer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/beevik/etree"
	"github.com/google/uuid"
	dsig "github.com/russellhaering/goxmldsig"
)

const (
	digestMethodSHA256       = "http://www.w3.org/2001/04/xmlenc#sha256"
	signatureMethodRSASHA256 = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
	envelopedSignatureAlgo   = "http://www.w3.org/2000/09/xmldsig#enveloped-signature"
	signedPropertiesRefType  = "http://uri.etsi.org/01903#SignedProperties"
)

// canonicalizer implements exactly "http://www.w3.org/TR/2001/REC-xml-c14n-20010315"
// (inclusive C14N 1.0) — the algorithm DIAN uses, distinct from the exclusive one
// (xml-exc-c14n).
var canonicalizer = dsig.MakeC14N10RecCanonicalizer()

// Signer signs UBL documents with a concrete RSA certificate and private key.
type Signer struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
}

// New creates a Signer from an already-loaded certificate and key (see LoadPEM/LoadPKCS12).
func New(cert *x509.Certificate, key *rsa.PrivateKey) *Signer {
	return &Signer{cert: cert, key: key}
}

// Sign signs root by inserting an XAdES-EPES ds:Signature inside placeholder.
//
// placeholder must be an empty ext:ExtensionContent already located in its final position
// within the document (the last ext:UBLExtension of ext:UBLExtensions, exactly as
// builder.BuildInvoice leaves it) — the signature needs to be in its final position for
// inclusive canonicalization (C14N 1.0 REC) to correctly inherit the document's namespaces.
// root is the document's root element (Invoice/CreditNote/DebitNote/AttachedDocument); its
// digest (ds:Reference URI="") is computed before the tree is touched further.
//
// role is the value of xades:ClaimedRole: "supplier" for Invoice/CreditNote/DebitNote, "" for
// the AttachedDocument's own signature (verified against both cases in real invoices).
func (s *Signer) Sign(root, placeholder *etree.Element, role string, signingTime time.Time) error {
	docDigest, err := digestValue(root)
	if err != nil {
		return fmt.Errorf("signer: document digest: %w", err)
	}

	id := "xmldsig-" + uuid.New().String()
	sigEl := placeholder.CreateElement("ds:Signature")
	sigEl.CreateAttr("Id", id)

	keyInfoEl := s.buildKeyInfo(sigEl, id)
	keyInfoDigest, err := digestValue(keyInfoEl)
	if err != nil {
		return fmt.Errorf("signer: KeyInfo digest: %w", err)
	}

	signedPropsEl := s.buildSignedProperties(sigEl, id, role, signingTime)
	signedPropsDigest, err := digestValue(signedPropsEl)
	if err != nil {
		return fmt.Errorf("signer: SignedProperties digest: %w", err)
	}

	signedInfoEl := buildSignedInfo(id, docDigest, keyInfoDigest, signedPropsDigest)
	sigEl.InsertChildAt(0, signedInfoEl) // schema order: SignedInfo, SignatureValue, KeyInfo, Object

	canonSignedInfo, err := canonicalizer.Canonicalize(signedInfoEl)
	if err != nil {
		return fmt.Errorf("signer: canonicalize SignedInfo: %w", err)
	}
	hashed := sha256.Sum256(canonSignedInfo)
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, hashed[:])
	if err != nil {
		return fmt.Errorf("signer: sign SignedInfo: %w", err)
	}

	sigValueEl := etree.NewElement("ds:SignatureValue")
	sigValueEl.CreateAttr("Id", id+"-sigvalue")
	sigValueEl.SetText(base64.StdEncoding.EncodeToString(signature))
	sigEl.InsertChildAt(1, sigValueEl)

	return nil
}

func buildSignedInfo(id, docDigest, keyInfoDigest, signedPropsDigest string) *etree.Element {
	signedInfo := etree.NewElement("ds:SignedInfo")
	signedInfo.CreateElement("ds:CanonicalizationMethod").CreateAttr("Algorithm", string(canonicalizer.Algorithm()))
	signedInfo.CreateElement("ds:SignatureMethod").CreateAttr("Algorithm", signatureMethodRSASHA256)

	ref0 := signedInfo.CreateElement("ds:Reference")
	ref0.CreateAttr("Id", id+"-ref0")
	ref0.CreateAttr("URI", "")
	ref0.CreateElement("ds:Transforms").CreateElement("ds:Transform").CreateAttr("Algorithm", envelopedSignatureAlgo)
	ref0.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", digestMethodSHA256)
	ref0.CreateElement("ds:DigestValue").SetText(docDigest)

	ref1 := signedInfo.CreateElement("ds:Reference")
	ref1.CreateAttr("Id", id+"-ref1")
	ref1.CreateAttr("URI", "#"+id+"-keyinfo")
	ref1.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", digestMethodSHA256)
	ref1.CreateElement("ds:DigestValue").SetText(keyInfoDigest)

	ref2 := signedInfo.CreateElement("ds:Reference")
	ref2.CreateAttr("Type", signedPropertiesRefType)
	ref2.CreateAttr("URI", "#"+id+"-signedprops")
	ref2.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", digestMethodSHA256)
	ref2.CreateElement("ds:DigestValue").SetText(signedPropsDigest)

	return signedInfo
}

// digestValue canonicalizes el (already positioned in the tree) and returns its SHA-256 in
// base64, ready for a ds:DigestValue.
func digestValue(el *etree.Element) (string, error) {
	canon, err := canonicalizer.Canonicalize(el)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canon)
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}
