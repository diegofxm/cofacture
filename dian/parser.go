// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
//
// Package dian interprets DIAN's validation responses (soap.DianResponse) and turns them into
// something the rest of the pipeline can use directly.
//
// Verified against a real GetStatusZip response: the XmlBase64Bytes field arrives as base64
// text that must be decoded to reach the actual ApplicationResponse — encoding/xml does not
// decode base64 on its own (it only copies an element's character data verbatim into a []byte
// field, confirmed directly against the standard library), so this one decode step is entirely
// this package's own doing, not something Go already did for us. XmlBytes, when DIAN uses it
// instead of XmlBase64Bytes, arrives already as the final content and needs no decode step.
package dian

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/diegofxm/cofacture/domain"
	"github.com/diegofxm/cofacture/soap"
)

// ValidatorID is the fixed literal DIAN uses as the validator identifier, confirmed both in
// the technical annex example and in a real response.
const ValidatorID = "Unidad Especial Dirección de Impuestos y Aduanas Nacionales"

// Message is a validation message already split into its parts. DIAN delivers them as a
// single string with the format "Regla: <code>, <Rechazo|Notificación>: <text>" — Severity is
// left empty if the message doesn't follow that pattern (it isn't discarded; Raw always keeps
// the original).
type Message struct {
	Rule     string
	Severity string
	Text     string
	Raw      string
}

// IsRejection is true when Severity is "Rechazo" — distinguishes an actual rejection from a
// plain informational notice (DIAN can approve a document while still attaching notices).
func (m Message) IsRejection() bool {
	return m.Severity == "Rechazo"
}

func parseMessage(raw string) Message {
	msg := Message{Raw: raw}
	rest, ok := strings.CutPrefix(raw, "Regla: ")
	if !ok {
		return msg
	}
	rule, rest, ok := strings.Cut(rest, ", ")
	if !ok {
		return msg
	}
	severity, text, ok := strings.Cut(rest, ": ")
	if !ok {
		return msg
	}
	msg.Rule = rule
	msg.Severity = severity
	msg.Text = text
	return msg
}

// Result is the interpretation of a soap.DianResponse: messages already split apart and, if
// present, the embedded ApplicationResponse already decoded.
type Result struct {
	IsValid           bool
	StatusCode        string
	StatusDescription string
	StatusMessage     string
	Messages          []Message

	// XmlDocumentKey is the identifier of the document this response applies to (the CUFE/CUDE
	// of the validated invoice/note), not the embedded ApplicationResponse's own identifier.
	XmlDocumentKey string
	XmlFileName    string

	// ApplicationResponseXML is the decoded XML of the ApplicationResponse DIAN issued during
	// validation, ready to use as domain.ValidationResult.ApplicationResponseXML.
	// nil if the response carried none.
	ApplicationResponseXML []byte
}

// HasRejections is true if any message is an actual rejection (not just a notice).
func (r Result) HasRejections() bool {
	for _, m := range r.Messages {
		if m.IsRejection() {
			return true
		}
	}
	return false
}

// IsTestSetClosed detects DIAN's specific response for when the test-set identifier (used in
// SendTestSetAsync) has already been certified/closed on their end — distinct from an actual
// rejection of the document's content, which arrives via ErrorMessage.Items (Messages), not
// here. Confirmed against a real response: no Messages, StatusCode "2", StatusDescription
// "Set de prueba con identificador <uuid> se encuentra Aceptado." — callers should retry via
// SendBillSync instead of treating this as a genuine document rejection.
func (r Result) IsTestSetClosed() bool {
	return strings.Contains(r.StatusDescription, "Set de prueba") &&
		strings.Contains(r.StatusDescription, "se encuentra Aceptado")
}

// Interpret converts a soap.DianResponse (one element of what GetStatusZip returns, or the
// result of GetStatus) into a Result.
func Interpret(resp soap.DianResponse) (Result, error) {
	result := Result{
		IsValid:           resp.IsValid,
		StatusCode:        resp.StatusCode,
		StatusDescription: resp.StatusDescription,
		StatusMessage:     resp.StatusMessage,
		XmlDocumentKey:    resp.XmlDocumentKey,
		XmlFileName:       resp.XmlFileName,
	}

	if resp.ErrorMessage != nil {
		for _, raw := range resp.ErrorMessage.Items {
			result.Messages = append(result.Messages, parseMessage(raw))
		}
	}

	embedded, err := decodeEmbeddedXML(resp)
	if err != nil {
		return Result{}, fmt.Errorf("dian: decode embedded XML: %w", err)
	}
	result.ApplicationResponseXML = embedded

	return result, nil
}

func decodeEmbeddedXML(resp soap.DianResponse) ([]byte, error) {
	if len(resp.XmlBase64Bytes) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(string(resp.XmlBase64Bytes))
		if err != nil {
			return nil, fmt.Errorf("decode XmlBase64Bytes: %w", err)
		}
		return decoded, nil
	}
	if len(resp.XmlBytes) > 0 {
		return resp.XmlBytes, nil
	}
	return nil, nil
}

// ToValidationResult builds a domain.ValidationResult ready for
// builder.BuildInvoiceAttachedDocument. DIAN's response does not repeat the document's own
// ID/date nor the moment the query was made, so the caller supplies them.
func (r Result) ToValidationResult(lineID, documentID, documentHashType, documentIssueDate, validationDate, validationTime string) domain.ValidationResult {
	return domain.ValidationResult{
		LineID:                 lineID,
		DocumentID:             documentID,
		DocumentCUFE:           r.XmlDocumentKey,
		DocumentHashType:       documentHashType,
		DocumentIssueDate:      documentIssueDate,
		ApplicationResponseXML: string(r.ApplicationResponseXML),
		ValidatorID:            ValidatorID,
		ValidationResultCode:   r.StatusCode,
		ValidationDate:         validationDate,
		ValidationTime:         validationTime,
	}
}
