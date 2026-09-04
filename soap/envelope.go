// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package soap implements the SOAP 1.2 client for DIAN's receiving web services
// (WcfDianCustomerServices), including the WS-Security header.
//
// Verified against the real certification-environment server (vpfe-hab.dian.gov.co): an
// invoice built with this client was sent via SendTestSetAsync and authorized
// (StatusCode=00, "ha sido autorizada").
//
// The policy published in the WSDL describes RequireThumbprintReference, but that variant
// produced "An error occurred when verifying security for the message" against the real
// server. The pattern that does work (matching existing real-world implementations: Chilkat,
// PHP's soap-dian) is different:
//
//   - An embedded wsse:BinarySecurityToken with a Direct Reference in KeyInfo, not by
//     thumbprint.
//   - Exclusive C14N forcing a fixed list of inclusive namespaces "wsa soap wcf" — without
//     this, the signature doesn't match what the server recomputes when verifying.
//   - TransportBinding with HttpsToken RequireClientCertificate="false" — plain HTTPS, no
//     mutual TLS, despite the technical annex (section 7.5) suggesting otherwise.
//   - Only the wsa:To header is signed (not the Body or the Timestamp).
//   - AlgorithmSuite Basic256Sha256Rsa15 — SHA-256 digest, RSA-SHA256 signature.
//   - Strict Layout — inside wsse:Security: Timestamp, BinarySecurityToken, Signature.
//   - Every element carrying a "wsu:Id" attribute must declare xmlns:wsu on itself — XML
//     namespaces declared on one element aren't inherited by its siblings. This caused an
//     empty 400 response (the request was rejected before reaching security processing) until
//     it was fixed.
package soap

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/beevik/etree"
	"github.com/google/uuid"
	dsig "github.com/russellhaering/goxmldsig"
)

const (
	nsSOAP12 = "http://www.w3.org/2003/05/soap-envelope"
	nsWSA    = "http://www.w3.org/2005/08/addressing"
	nsWCF    = "http://wcf.dian.colombia"
	nsWSSE   = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	nsWSU    = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
	nsDS     = "http://www.w3.org/2000/09/xmldsig#"
	nsEC14N  = "http://www.w3.org/2001/10/xml-exc-c14n#"

	actionBase = nsWCF + "/IWcfDianCustomerServices/"

	x509v3ValueType       = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3"
	base64EncodingType    = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"
	signatureMethodRSA256 = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
	digestMethodSHA256    = "http://www.w3.org/2001/04/xmlenc#sha256"

	// inclusiveNSPrefixList forces these prefixes to be declared in canonical form even though
	// the signed element doesn't "visibly use" them in the strict sense of the algorithm —
	// this is exactly what EXC-C14N's InclusiveNamespaces mechanism exists for, and it's what
	// real implementations against this same service use.
	inclusiveNSPrefixList = "wsa soap wcf"
)

// exclusiveCanonicalizer implements "http://www.w3.org/2001/10/xml-exc-c14n#" — the algorithm
// WS-Security requires, distinct from the inclusive C14N the document's XAdES signature uses
// (package signer). They intentionally don't share a canonicalizer: these are different
// signature layers.
var exclusiveCanonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(inclusiveNSPrefixList)

// buildEnvelope builds the full SOAP 1.2 envelope: WS-Addressing + WS-Security (signing only
// wsa:To) + the body bodyBuilder provides.
func (c *Client) buildEnvelope(action string, bodyBuilder func(body *etree.Element)) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	env := doc.CreateElement("soap:Envelope")
	env.CreateAttr("xmlns:soap", nsSOAP12)
	env.CreateAttr("xmlns:wcf", nsWCF)

	header := env.CreateElement("soap:Header")

	actionEl := header.CreateElement("wsa:Action")
	actionEl.CreateAttr("xmlns:wsa", nsWSA)
	actionEl.SetText(actionBase + action)

	const toID = "_to"
	toEl := header.CreateElement("wsa:To")
	toEl.CreateAttr("xmlns:wsa", nsWSA)
	toEl.CreateAttr("xmlns:wsu", nsWSU)
	// xmlns:soap and xmlns:wcf are declared here too — redundant for wsa:To's own content, but
	// needed so inclusiveNSPrefixList has something to retain when canonicalizing this element
	// in isolation (see the note on exclusiveCanonicalizer above).
	toEl.CreateAttr("xmlns:soap", nsSOAP12)
	toEl.CreateAttr("xmlns:wcf", nsWCF)
	toEl.CreateAttr("wsu:Id", toID)
	toEl.SetText(c.BaseURL)

	replyTo := header.CreateElement("wsa:ReplyTo")
	replyTo.CreateAttr("xmlns:wsa", nsWSA)
	replyTo.CreateElement("wsa:Address").SetText(nsWSA + "/anonymous")

	msgID := header.CreateElement("wsa:MessageID")
	msgID.CreateAttr("xmlns:wsa", nsWSA)
	msgID.SetText("urn:uuid:" + uuid.New().String())

	if err := c.appendSecurityHeader(header, toEl, toID); err != nil {
		return nil, fmt.Errorf("soap: build wsse:Security: %w", err)
	}

	body := env.CreateElement("soap:Body")
	bodyBuilder(body)

	return doc, nil
}

func (c *Client) appendSecurityHeader(header, toEl *etree.Element, toID string) error {
	sec := header.CreateElement("wsse:Security")
	sec.CreateAttr("xmlns:wsse", nsWSSE)
	sec.CreateAttr("soap:mustUnderstand", "1")

	now := time.Now().UTC()
	ts := sec.CreateElement("wsu:Timestamp")
	ts.CreateAttr("xmlns:wsu", nsWSU)
	ts.CreateAttr("wsu:Id", "_ts")
	ts.CreateElement("wsu:Created").SetText(now.Format("2006-01-02T15:04:05.000Z"))
	ts.CreateElement("wsu:Expires").SetText(now.Add(5 * time.Minute).Format("2006-01-02T15:04:05.000Z"))

	tokenID := "X509-" + uuid.New().String()
	bst := sec.CreateElement("wsse:BinarySecurityToken")
	// xmlns:wsu must be declared here: it isn't inherited from the sibling Timestamp element
	// (XML namespace declarations only flow down to descendants, never across siblings).
	// Without this, "wsu:Id" uses an undeclared prefix and a strict namespace parser (like
	// WCF's) rejects the document before it even processes security — exactly the empty 400
	// the real server used to return.
	bst.CreateAttr("xmlns:wsu", nsWSU)
	bst.CreateAttr("wsu:Id", tokenID)
	bst.CreateAttr("EncodingType", base64EncodingType)
	bst.CreateAttr("ValueType", x509v3ValueType)
	bst.SetText(base64.StdEncoding.EncodeToString(c.Cert.Raw))

	return c.appendSignature(sec, toEl, toID, tokenID)
}

// appendSignature signs only toEl (the wsa:To header, via Reference URI="#"+toID).
func (c *Client) appendSignature(sec, toEl *etree.Element, toID, tokenID string) error {
	canonTo, err := exclusiveCanonicalizer.Canonicalize(toEl)
	if err != nil {
		return err
	}
	digestTo := sha256.Sum256(canonTo)

	sigEl := sec.CreateElement("ds:Signature")
	sigEl.CreateAttr("xmlns:ds", nsDS)

	signedInfo := sigEl.CreateElement("ds:SignedInfo")
	// The exclusive canonicalizer (unlike the inclusive one package signer uses for XAdES)
	// does not inherit namespaces from ancestors: it treats the canonicalized element as if it
	// were the root. xmlns:ds/soap/wcf must be declared within this subtree.
	signedInfo.CreateAttr("xmlns:ds", nsDS)
	signedInfo.CreateAttr("xmlns:soap", nsSOAP12)
	signedInfo.CreateAttr("xmlns:wcf", nsWCF)

	canonMethod := signedInfo.CreateElement("ds:CanonicalizationMethod")
	canonMethod.CreateAttr("Algorithm", nsEC14N)
	appendInclusiveNamespaces(canonMethod)

	signedInfo.CreateElement("ds:SignatureMethod").CreateAttr("Algorithm", signatureMethodRSA256)

	ref := signedInfo.CreateElement("ds:Reference")
	ref.CreateAttr("URI", "#"+toID)
	transform := ref.CreateElement("ds:Transforms").CreateElement("ds:Transform")
	transform.CreateAttr("Algorithm", nsEC14N)
	appendInclusiveNamespaces(transform)
	ref.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", digestMethodSHA256)
	ref.CreateElement("ds:DigestValue").SetText(base64.StdEncoding.EncodeToString(digestTo[:]))

	canonSignedInfo, err := exclusiveCanonicalizer.Canonicalize(signedInfo)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(canonSignedInfo)
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.Key, crypto.SHA256, hashed[:])
	if err != nil {
		return err
	}
	sigEl.CreateElement("ds:SignatureValue").SetText(base64.StdEncoding.EncodeToString(signature))

	keyInfo := sigEl.CreateElement("ds:KeyInfo")
	str := keyInfo.CreateElement("wsse:SecurityTokenReference")
	str.CreateAttr("xmlns:wsse", nsWSSE)
	tokenRef := str.CreateElement("wsse:Reference")
	tokenRef.CreateAttr("URI", "#"+tokenID)
	tokenRef.CreateAttr("ValueType", x509v3ValueType)

	return nil
}

func appendInclusiveNamespaces(transformOrMethod *etree.Element) {
	ec := transformOrMethod.CreateElement("ec:InclusiveNamespaces")
	ec.CreateAttr("xmlns:ec", nsEC14N)
	ec.CreateAttr("PrefixList", inclusiveNSPrefixList)
}
