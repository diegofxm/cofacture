// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package soap

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
)

// Real service URLs (verified against the downloaded WSDL directly, not inferred from the
// technical annex).
const (
	HabilitacionURL = "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc"
	ProduccionURL   = "https://vpfe.dian.gov.co/WcfDianCustomerServices.svc"
)

// Client is the SOAP client for DIAN's receiving web services. It does not require mutual TLS
// (the WSDL's real policy has RequireClientCertificate="false") — Cert/Key are used only to
// sign the wsa:To WS-Security header.
type Client struct {
	BaseURL    string
	Cert       *x509.Certificate
	Key        *rsa.PrivateKey
	HTTPClient *http.Client
}

// New creates a Client. baseURL is normally HabilitacionURL or ProduccionURL.
func New(baseURL string, cert *x509.Certificate, key *rsa.PrivateKey) *Client {
	return &Client{
		BaseURL:    baseURL,
		Cert:       cert,
		Key:        key,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Fault is a soap:Fault returned by DIAN.
type Fault struct {
	Code   string
	Reason string
}

func (f *Fault) Error() string {
	return fmt.Sprintf("soap fault [%s]: %s", f.Code, f.Reason)
}

type soapFaultXML struct {
	Code struct {
		Value string `xml:"Value"`
	} `xml:"Code"`
	Reason struct {
		Text string `xml:"Text"`
	} `xml:"Reason"`
}

// call sends action with the body bodyBuilder constructs, and leaves the raw response (the
// content of soap:Body, already verified not to be a Fault) in out via xml.Unmarshal.
// out must be a pointer to a struct with the xml tag matching the "{action}Response" element
// expected inside soap:Body.
func (c *Client) call(action string, bodyBuilder func(body *etree.Element), out any) error {
	doc, err := c.buildEnvelope(action, bodyBuilder)
	if err != nil {
		return err
	}
	reqXML, err := doc.WriteToBytes()
	if err != nil {
		return fmt.Errorf("soap: serialize request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.BaseURL, bytes.NewReader(reqXML))
	if err != nil {
		return fmt.Errorf("soap: build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", fmt.Sprintf(`application/soap+xml; charset=utf-8; action="%s%s"`, actionBase, action))

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("soap: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("soap: read response: %w", err)
	}

	var envelope struct {
		Body struct {
			Fault   *soapFaultXML `xml:"http://www.w3.org/2003/05/soap-envelope Fault"`
			Content []byte        `xml:",innerxml"`
		} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
	}
	if err := xml.Unmarshal(respBody, &envelope); err != nil {
		return fmt.Errorf("soap: parse response (HTTP %d, headers %v, %d bytes): %w\n%s", resp.StatusCode, resp.Header, len(respBody), err, respBody)
	}
	if envelope.Body.Fault != nil {
		return &Fault{Code: envelope.Body.Fault.Code.Value, Reason: envelope.Body.Fault.Reason.Text}
	}

	// The HTTP status is NOT a reliable success/error signal for this service — confirmed
	// against the real DIAN environment: GetAcquirer responds with HTTP 404 and a perfectly
	// valid SOAP body when the acquirer doesn't exist (StatusCode "404" INSIDE the body,
	// Message "El adquirente No existe en la base de datos", no soap:Fault) — that is the
	// normal, expected result for most identification numbers, not a transport error. That's
	// why the body is parsed into out FIRST, regardless of status; the status is only used to
	// give a clearer error message if the body isn't the expected response.
	if err := xml.Unmarshal(envelope.Body.Content, out); err != nil {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("soap: HTTP %d with no explicit soap:Fault:\n%s", resp.StatusCode, respBody)
		}
		return fmt.Errorf("soap: parse %sResponse: %w", action, err)
	}
	return nil
}
