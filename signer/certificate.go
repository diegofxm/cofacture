// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package signer

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"software.sslmate.com/src/go-pkcs12"
)

// LoadPEM loads a certificate and its RSA private key from PEM bytes (typically two separate
// files: *_cert.pem and *_key.pem, extracted from a .p12).
func LoadPEM(certPEM, keyPEM []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, errors.New("signer: no PEM certificate block found")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("signer: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, errors.New("signer: no PEM private key block found")
	}
	key, err := parsePrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

// LoadPKCS12 loads a certificate and its RSA private key from a .p12/.pfx file, exactly as
// issued by the certificate authority.
func LoadPKCS12(data []byte, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, cert, err := pkcs12.Decode(data, password)
	if err != nil {
		return nil, nil, fmt.Errorf("signer: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("signer: the .p12's private key is not RSA")
	}
	return cert, rsaKey, nil
}

func parsePrivateKey(der []byte) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("signer: the PKCS8 private key is not RSA")
		}
		return rsaKey, nil
	}
	key, err := x509.ParsePKCS1PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("signer: %w", err)
	}
	return key, nil
}
