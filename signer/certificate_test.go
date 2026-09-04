// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package signer

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"software.sslmate.com/src/go-pkcs12"
)

const testP12Password = "test-password"

// TestLoadPKCS12_WithCAChain confirms LoadPKCS12 accepts a .p12 that bundles a CA certificate
// alongside the leaf cert + key — the real-world shape DIAN-issued certificates routinely use.
// pkcs12.Decode rejects this ("expected exactly two safe bags in the PFX PDU"); LoadPKCS12 uses
// pkcs12.DecodeChain specifically to accept it. This is the regression test the CA-chain fix
// itself never got.
func TestLoadPKCS12_WithCAChain(t *testing.T) {
	leafCert, key := generateTestCert(t)
	caCert, _ := generateTestCert(t)

	data, err := pkcs12.Encode(rand.Reader, key, leafCert, []*x509.Certificate{caCert}, testP12Password)
	if err != nil {
		t.Fatalf("pkcs12.Encode: %v", err)
	}

	gotCert, gotKey, err := LoadPKCS12(data, testP12Password)
	if err != nil {
		t.Fatalf("LoadPKCS12: %v", err)
	}
	if !gotCert.Equal(leafCert) {
		t.Error("LoadPKCS12 returned a different certificate than the leaf that was encoded")
	}
	if gotKey.D.Cmp(key.D) != 0 {
		t.Error("LoadPKCS12 returned a different private key than the one that was encoded")
	}
}

// TestLoadPKCS12_NoChain confirms the simple case (leaf cert + key only, no CA bundled) still
// works — this is what pkcs12.Decode already handled before the CA-chain fix.
func TestLoadPKCS12_NoChain(t *testing.T) {
	cert, key := generateTestCert(t)

	data, err := pkcs12.Encode(rand.Reader, key, cert, nil, testP12Password)
	if err != nil {
		t.Fatalf("pkcs12.Encode: %v", err)
	}

	gotCert, _, err := LoadPKCS12(data, testP12Password)
	if err != nil {
		t.Fatalf("LoadPKCS12: %v", err)
	}
	if !gotCert.Equal(cert) {
		t.Error("LoadPKCS12 returned a different certificate than the one that was encoded")
	}
}

// TestLoadPKCS12_WrongPassword confirms a wrong password is reported as an error.
func TestLoadPKCS12_WrongPassword(t *testing.T) {
	cert, key := generateTestCert(t)
	data, err := pkcs12.Encode(rand.Reader, key, cert, nil, testP12Password)
	if err != nil {
		t.Fatalf("pkcs12.Encode: %v", err)
	}

	if _, _, err := LoadPKCS12(data, "wrong-password"); err == nil {
		t.Error("LoadPKCS12 should fail with the wrong password")
	}
}

// TestLoadPEM_RoundTrip confirms LoadPEM reconstructs the same certificate and key from their
// PEM-encoded forms (the shape *_cert.pem/*_key.pem files take, per the package doc comment).
func TestLoadPEM_RoundTrip(t *testing.T) {
	cert, key := generateTestCert(t)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	gotCert, gotKey, err := LoadPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadPEM: %v", err)
	}
	if !gotCert.Equal(cert) {
		t.Error("LoadPEM returned a different certificate than the one that was encoded")
	}
	if gotKey.D.Cmp(key.D) != 0 {
		t.Error("LoadPEM returned a different private key than the one that was encoded")
	}
}

// TestLoadPEM_InvalidCertBlock confirms a missing/invalid certificate PEM block is reported as
// an error instead of panicking on a nil block.
func TestLoadPEM_InvalidCertBlock(t *testing.T) {
	_, key := generateTestCert(t)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	if _, _, err := LoadPEM([]byte("not a pem block"), keyPEM); err == nil {
		t.Error("LoadPEM should fail when the certificate PEM block is missing/invalid")
	}
}
