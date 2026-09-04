// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package soap

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"github.com/beevik/etree"
)

func generateTestCert(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return cert, key
}

// TestBuildEnvelope_SignatureVerifies builds a complete envelope and verifies it the same way
// an independent verifier would: canonicalize wsa:To and ds:SignedInfo with exclusive C14N
// (not the inclusive variant used by XAdES) and check the signature against the public key.
func TestBuildEnvelope_SignatureVerifies(t *testing.T) {
	cert, key := generateTestCert(t)
	c := New(HabilitacionURL, cert, key)

	doc, err := c.buildEnvelope("SendTestSetAsync", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendTestSetAsync")
		el.CreateElement("wcf:fileName").SetText("z0000000000000190000000B.zip")
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString([]byte("test content")))
		el.CreateElement("wcf:testSetId").SetText("653bf9d9-b2b1-44ae-a66d-3b9cdc4271c3")
	})
	if err != nil {
		t.Fatalf("buildEnvelope: %v", err)
	}

	root := doc.Root()

	toEl := root.FindElement("//wsa:To")
	if toEl == nil {
		t.Fatal("wsa:To not found")
	}
	if toEl.Text() != HabilitacionURL {
		t.Errorf("wsa:To = %q, want %q", toEl.Text(), HabilitacionURL)
	}

	sigEl := root.FindElement("//ds:Signature")
	if sigEl == nil {
		t.Fatal("ds:Signature not found")
	}
	signedInfo := sigEl.FindElement("ds:SignedInfo")
	if signedInfo == nil {
		t.Fatal("ds:Signature has no ds:SignedInfo")
	}

	// The reference must point only to the signed wsa:To (Body and Timestamp are NOT
	// signed — that's what sp:SignedParts requires in the WSDL's real policy).
	refs := signedInfo.SelectElements("ds:Reference")
	if len(refs) != 1 {
		t.Fatalf("expected exactly 1 ds:Reference (only wsa:To), got %d", len(refs))
	}
	if uri := refs[0].SelectAttrValue("URI", ""); uri != "#_to" {
		t.Errorf("Reference URI = %q, want %q", uri, "#_to")
	}

	canonSignedInfo, err := exclusiveCanonicalizer.Canonicalize(signedInfo)
	if err != nil {
		t.Fatalf("canonicalize SignedInfo: %v", err)
	}
	hashed := sha256.Sum256(canonSignedInfo)

	sigValueEl := sigEl.FindElement("ds:SignatureValue")
	if sigValueEl == nil {
		t.Fatal("ds:SignatureValue not found")
	}
	sigValue, err := base64.StdEncoding.DecodeString(sigValueEl.Text())
	if err != nil {
		t.Fatalf("decode SignatureValue: %v", err)
	}
	if err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hashed[:], sigValue); err != nil {
		t.Errorf("signature does not verify against the public key: %v", err)
	}

	// The certificate is embedded as a BinarySecurityToken, referenced by Direct
	// Reference from KeyInfo — not by thumbprint (see the package comment: the
	// thumbprint variant the WSDL's published policy asks for was rejected by the real server).
	bst := root.FindElement("//wsse:BinarySecurityToken")
	if bst == nil {
		t.Fatal("wsse:BinarySecurityToken not found")
	}
	if got := bst.Text(); got != base64.StdEncoding.EncodeToString(cert.Raw) {
		t.Error("BinarySecurityToken does not contain the expected certificate")
	}
	tokenID := bst.SelectAttrValue("wsu:Id", "")
	if tokenID == "" {
		t.Fatal("BinarySecurityToken has no wsu:Id")
	}

	tokenRef := sigEl.FindElement("ds:KeyInfo/wsse:SecurityTokenReference/wsse:Reference")
	if tokenRef == nil {
		t.Fatal("wsse:Reference not found inside KeyInfo")
	}
	if uri := tokenRef.SelectAttrValue("URI", ""); uri != "#"+tokenID {
		t.Errorf("wsse:Reference URI = %q, want %q", uri, "#"+tokenID)
	}
}
