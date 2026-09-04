// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package dian

import (
	"encoding/base64"
	"testing"

	"github.com/diegofxm/cofacture/soap"
)

func TestParseMessage(t *testing.T) {
	cases := []struct {
		raw  string
		want Message
	}{
		// Real messages returned by DIAN (certification/testing environment).
		{
			raw: "Regla: ZE02, Rechazo: Valor de la firma inválido.",
			want: Message{
				Rule: "ZE02", Severity: "Rechazo", Text: "Valor de la firma inválido.",
				Raw: "Regla: ZE02, Rechazo: Valor de la firma inválido.",
			},
		},
		{
			raw: "Regla: FAJ43b, Notificación: Nombre informado No corresponde al registrado en el RUT con respecto al Nit suminstrado.",
			want: Message{
				Rule: "FAJ43b", Severity: "Notificación",
				Text: "Nombre informado No corresponde al registrado en el RUT con respecto al Nit suminstrado.",
				Raw:  "Regla: FAJ43b, Notificación: Nombre informado No corresponde al registrado en el RUT con respecto al Nit suminstrado.",
			},
		},
		{
			// Text that doesn't follow the pattern: it is preserved in Raw, the rest is left
			// empty instead of failing.
			raw:  "algo que no sigue el formato esperado",
			want: Message{Raw: "algo que no sigue el formato esperado"},
		},
	}

	for _, c := range cases {
		got := parseMessage(c.raw)
		if got != c.want {
			t.Errorf("parseMessage(%q) = %+v, want %+v", c.raw, got, c.want)
		}
	}
}

func TestMessage_IsRejection(t *testing.T) {
	if !(Message{Severity: "Rechazo"}).IsRejection() {
		t.Error(`Severity "Rechazo" should count as a rejection`)
	}
	if (Message{Severity: "Notificación"}).IsRejection() {
		t.Error(`Severity "Notificación" should not count as a rejection`)
	}
}

func TestInterpret_DecodesBase64ApplicationResponse(t *testing.T) {
	// XmlBase64Bytes arrives as base64 text — encoding/xml never decodes it on its own (it
	// only copies character data verbatim into a []byte field), so decodeEmbeddedXML's one
	// base64.StdEncoding.DecodeString call is the only decode step, not a "second layer" of
	// one Go already did — verified against a real GetStatusZip response.
	innerXML := `<?xml version="1.0" encoding="utf-8"?><ApplicationResponse><cbc:UUID>cude-de-prueba</cbc:UUID></ApplicationResponse>`
	encoded := base64.StdEncoding.EncodeToString([]byte(innerXML))

	resp := soap.DianResponse{
		IsValid:        true,
		StatusCode:     "00",
		StatusMessage:  "ha sido autorizada",
		XmlDocumentKey: "cufe-de-prueba",
		XmlBase64Bytes: []byte(encoded),
	}

	result, err := Interpret(resp)
	if err != nil {
		t.Fatalf("Interpret: %v", err)
	}
	if string(result.ApplicationResponseXML) != innerXML {
		t.Errorf("ApplicationResponseXML = %q, want %q", result.ApplicationResponseXML, innerXML)
	}
	if !result.IsValid || result.StatusCode != "00" {
		t.Errorf("basic fields were not copied correctly: %+v", result)
	}
}

func TestInterpret_NoEmbeddedXML(t *testing.T) {
	result, err := Interpret(soap.DianResponse{IsValid: false, StatusCode: "99"})
	if err != nil {
		t.Fatalf("Interpret: %v", err)
	}
	if result.ApplicationResponseXML != nil {
		t.Error("ApplicationResponseXML should be nil when the response carries none")
	}
}

func TestResult_HasRejectionsAndToValidationResult(t *testing.T) {
	result := Result{
		IsValid:        true,
		StatusCode:     "00",
		XmlDocumentKey: "cufe-123",
		Messages: []Message{
			parseMessage("Regla: FAJ43b, Notificación: algo informativo."),
		},
		ApplicationResponseXML: []byte("<ApplicationResponse/>"),
	}
	if result.HasRejections() {
		t.Error("a lone notice should not count as a rejection")
	}

	result.Messages = append(result.Messages, parseMessage("Regla: ZE02, Rechazo: firma inválida."))
	if !result.HasRejections() {
		t.Error("should detect the rejection that was appended")
	}

	vr := result.ToValidationResult("1", "SETP1", "CUFE-SHA384", "2024-01-20", "2024-01-20", "10:00:00-05:00")
	if vr.DocumentCUFE != "cufe-123" {
		t.Errorf("DocumentCUFE = %q, want %q", vr.DocumentCUFE, "cufe-123")
	}
	if vr.ValidatorID != ValidatorID {
		t.Errorf("ValidatorID = %q, want %q", vr.ValidatorID, ValidatorID)
	}
	if vr.ApplicationResponseXML != "<ApplicationResponse/>" {
		t.Errorf("ApplicationResponseXML was not propagated: %q", vr.ApplicationResponseXML)
	}
}

func TestResult_IsTestSetClosed(t *testing.T) {
	// Real response that produced the rejection in section 9.43 — no Messages, just
	// StatusDescription.
	closed := Result{
		StatusCode:        "2",
		StatusDescription: "Set de prueba con identificador 653bf9d9-b2b1-44ae-a66d-3b9cdc4271c3 se encuentra Aceptado.",
	}
	if !closed.IsTestSetClosed() {
		t.Error("should detect the closed test set")
	}

	// A normal content rejection must not be confused with this, even though it also has
	// StatusCode "2".
	contentRejection := Result{
		StatusCode:        "2",
		StatusDescription: "Documento rechazado",
		Messages:          []Message{parseMessage("Regla: ZE02, Rechazo: Valor de la firma inválido.")},
	}
	if contentRejection.IsTestSetClosed() {
		t.Error("a normal content rejection should not be detected as a closed test set")
	}
}
