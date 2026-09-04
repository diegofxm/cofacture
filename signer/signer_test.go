// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package signer

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
	"github.com/diegofxm/cofacture/domain"
)

func generateTestCert(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generar llave: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("crear certificado: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parsear certificado: %v", err)
	}
	return cert, key
}

// newTestDoc builds the minimal skeleton of a UBL document: namespaces on the root and the
// signature placeholder in its final position, exactly as builder.BuildInvoice leaves it.
func newTestDoc() (root, placeholder *etree.Element) {
	doc := etree.NewDocument()
	root = doc.CreateElement("Invoice")
	root.CreateAttr("xmlns", "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2")
	root.CreateAttr("xmlns:cac", "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2")
	root.CreateAttr("xmlns:cbc", "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2")
	root.CreateAttr("xmlns:ds", "http://www.w3.org/2000/09/xmldsig#")
	root.CreateAttr("xmlns:ext", "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2")
	root.CreateAttr("xmlns:xades", "http://uri.etsi.org/01903/v1.3.2#")
	extensions := root.CreateElement("ext:UBLExtensions")
	placeholder = extensions.CreateElement("ext:UBLExtension").CreateElement("ext:ExtensionContent")
	root.CreateElement("cbc:ID").SetText("TEST1")
	return root, placeholder
}

// verifySignature reconstructs what an independent verifier would do: canonicalize
// ds:SignedInfo and check ds:SignatureValue against the public key.
func verifySignature(t *testing.T, root *etree.Element, pub *rsa.PublicKey) {
	t.Helper()
	sig := root.FindElement("//ds:Signature")
	if sig == nil {
		t.Fatal("no se insertó ds:Signature")
	}

	signedInfo := sig.FindElement("ds:SignedInfo")
	if signedInfo == nil {
		t.Fatal("ds:Signature no tiene ds:SignedInfo")
	}
	canon, err := canonicalizer.Canonicalize(signedInfo)
	if err != nil {
		t.Fatalf("canonicalizar SignedInfo: %v", err)
	}
	hashed := sha256.Sum256(canon)

	sigValueEl := sig.FindElement("ds:SignatureValue")
	if sigValueEl == nil {
		t.Fatal("ds:Signature no tiene ds:SignatureValue")
	}
	sigValue, err := base64.StdEncoding.DecodeString(sigValueEl.Text())
	if err != nil {
		t.Fatalf("decodificar SignatureValue: %v", err)
	}

	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sigValue); err != nil {
		t.Errorf("signature does not verify against the public key: %v", err)
	}

	refs := signedInfo.SelectElements("ds:Reference")
	if len(refs) != 3 {
		t.Fatalf("se esperaban 3 ds:Reference (documento, KeyInfo, SignedProperties), hay %d", len(refs))
	}
}

func TestSign_RoundTrip(t *testing.T) {
	cert, key := generateTestCert(t)
	root, placeholder := newTestDoc()

	s := New(cert, key)
	if err := s.Sign(root, placeholder, "supplier", time.Now().In(domain.Bogota)); err != nil {
		t.Fatalf("Sign: %v", err)
	}

	verifySignature(t, root, &key.PublicKey)
}

func TestSign_EmptyRoleOmitsSignerRole(t *testing.T) {
	cert, key := generateTestCert(t)
	root, placeholder := newTestDoc()

	s := New(cert, key)
	if err := s.Sign(root, placeholder, "", time.Now().In(domain.Bogota)); err != nil {
		t.Fatalf("Sign: %v", err)
	}

	if root.FindElement("//xades:SignerRole") != nil {
		t.Error("no debería existir xades:SignerRole cuando role es vacío (caso AttachedDocument)")
	}
	verifySignature(t, root, &key.PublicKey)
}
