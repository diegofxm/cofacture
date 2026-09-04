// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package soap

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// canned simulates a complete SOAP response from the DIAN — the real client
// (TestBuildEnvelope_SignatureVerifies in envelope_test.go) already proves that the REQUEST is
// signed correctly; these tests cover the other side: that each response is parsed correctly.

const getNumberingRangeResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetNumberingRangeResponse xmlns="http://wcf.dian.colombia">
      <GetNumberingRangeResult xmlns:b="http://schemas.datacontract.org/2004/07/NumberRangeResponseList"
                               xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:OperationCode>0</b:OperationCode>
        <b:OperationDescription>Proceso Exitoso</b:OperationDescription>
        <b:ResponseList xmlns:c="http://schemas.datacontract.org/2004/07/NumberRangeResponse">
          <c:NumberRangeResponse>
            <c:ResolutionNumber>18760000001</c:ResolutionNumber>
            <c:ResolutionDate>2019-01-19T00:00:00</c:ResolutionDate>
            <c:Prefix>SETP</c:Prefix>
            <c:FromNumber>990000000</c:FromNumber>
            <c:ToNumber>995000000</c:ToNumber>
            <c:ValidDateFrom>2019-01-19T00:00:00</c:ValidDateFrom>
            <c:ValidDateTo>2030-01-19T00:00:00</c:ValidDateTo>
            <c:TechnicalKey>fc8eac422eba16e22ffd8c6f94b3f40a6e38162c</c:TechnicalKey>
          </c:NumberRangeResponse>
        </b:ResponseList>
      </GetNumberingRangeResult>
    </GetNumberingRangeResponse>
  </s:Body>
</s:Envelope>`

// TestGetNumberingRange_ParsesResponse confirms that GetNumberingRange correctly parses a
// response with one range — including the fields the CUFE needs (TechnicalKey) and the fields
// needed to record the range in the DB (Prefix, From/To, dates, resolution).
func TestGetNumberingRange_ParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getNumberingRangeResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetNumberingRange("6382356", "6382356", "12345678-1234-1234-1234-123456789012")
	if err != nil {
		t.Fatalf("GetNumberingRange: %v", err)
	}
	if got.OperationCode != "0" {
		t.Errorf("OperationCode = %q, want %q", got.OperationCode, "0")
	}
	if len(got.ResponseList) != 1 {
		t.Fatalf("len(ResponseList) = %d, want 1", len(got.ResponseList))
	}
	r := got.ResponseList[0]
	if r.ResolutionNumber != "18760000001" {
		t.Errorf("ResolutionNumber = %q, want %q", r.ResolutionNumber, "18760000001")
	}
	if r.Prefix != "SETP" {
		t.Errorf("Prefix = %q, want %q", r.Prefix, "SETP")
	}
	if r.FromNumber != 990000000 {
		t.Errorf("FromNumber = %d, want %d", r.FromNumber, 990000000)
	}
	if r.ToNumber != 995000000 {
		t.Errorf("ToNumber = %d, want %d", r.ToNumber, 995000000)
	}
	if r.TechnicalKey != "fc8eac422eba16e22ffd8c6f94b3f40a6e38162c" {
		t.Errorf("TechnicalKey = %q, want %q", r.TechnicalKey, "fc8eac422eba16e22ffd8c6f94b3f40a6e38162c")
	}
}

// TestGetNumberingRange_EmptyList confirms that a response with an empty list doesn't fail —
// this can happen when the software has no authorized ranges yet or the credentials are new.
func TestGetNumberingRange_EmptyList(t *testing.T) {
	emptyXML := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetNumberingRangeResponse xmlns="http://wcf.dian.colombia">
      <GetNumberingRangeResult xmlns:b="http://schemas.datacontract.org/2004/07/NumberRangeResponseList">
        <b:OperationCode>0</b:OperationCode>
        <b:OperationDescription>Proceso Exitoso</b:OperationDescription>
        <b:ResponseList/>
      </GetNumberingRangeResult>
    </GetNumberingRangeResponse>
  </s:Body>
</s:Envelope>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(emptyXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetNumberingRange("6382356", "6382356", "any-software-id")
	if err != nil {
		t.Fatalf("GetNumberingRange with empty list: %v", err)
	}
	if len(got.ResponseList) != 0 {
		t.Errorf("len(ResponseList) = %d, want 0", len(got.ResponseList))
	}
}

const sendBillAttachmentAsyncResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <SendBillAttachmentAsyncResponse xmlns="http://wcf.dian.colombia">
      <SendBillAttachmentAsyncResult xmlns:b="http://schemas.datacontract.org/2004/07/"
                                      xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:ZipKey>a1b2c3d4-e5f6-7890-abcd-ef1234567890</b:ZipKey>
      </SendBillAttachmentAsyncResult>
    </SendBillAttachmentAsyncResponse>
  </s:Body>
</s:Envelope>`

// TestSendBillAttachmentAsync_ParsesResponse confirms SendBillAttachmentAsync sends the same
// fileName/contentFile request shape as SendBillAsync and correctly parses the ZipKey out of
// its own distinct <SendBillAttachmentAsyncResult> response element.
func TestSendBillAttachmentAsync_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(sendBillAttachmentAsyncResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.SendBillAttachmentAsync("ad900123456000123456789.zip", []byte("<AttachedDocument/>"))
	if err != nil {
		t.Fatalf("SendBillAttachmentAsync: %v", err)
	}
	if got.ZipKey != "a1b2c3d4-e5f6-7890-abcd-ef1234567890" {
		t.Errorf("ZipKey = %q, want %q", got.ZipKey, "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	}
	if !strings.Contains(string(gotBody), "<wcf:SendBillAttachmentAsync>") {
		t.Errorf("request body missing wcf:SendBillAttachmentAsync element: %s", gotBody)
	}
	if !strings.Contains(string(gotBody), "ad900123456000123456789.zip") {
		t.Errorf("request body missing fileName: %s", gotBody)
	}
}

const sendEventUpdateStatusResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <SendEventUpdateStatusResponse xmlns="http://wcf.dian.colombia">
      <SendEventUpdateStatusResult xmlns:b="http://schemas.datacontract.org/2004/07/"
                                    xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:IsValid>true</b:IsValid>
        <b:StatusCode>00</b:StatusCode>
        <b:StatusDescription>La Notificación ha sido autorizada</b:StatusDescription>
      </SendEventUpdateStatusResult>
    </SendEventUpdateStatusResponse>
  </s:Body>
</s:Envelope>`

// TestSendEventUpdateStatus_ParsesResponse confirms SendEventUpdateStatus sends the single
// contentFile parameter (no fileName, same shape as SendNominaSync) and correctly parses its
// own distinct <SendEventUpdateStatusResult> response element.
func TestSendEventUpdateStatus_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(sendEventUpdateStatusResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.SendEventUpdateStatus([]byte("<ApplicationResponse/>"))
	if err != nil {
		t.Fatalf("SendEventUpdateStatus: %v", err)
	}
	if !got.IsValid {
		t.Errorf("IsValid = %v, want true", got.IsValid)
	}
	if got.StatusCode != "00" {
		t.Errorf("StatusCode = %q, want %q", got.StatusCode, "00")
	}
	if !strings.Contains(string(gotBody), "<wcf:SendEventUpdateStatus>") {
		t.Errorf("request body missing wcf:SendEventUpdateStatus element: %s", gotBody)
	}
}

const getStatusEventResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetStatusEventResponse xmlns="http://wcf.dian.colombia">
      <GetStatusEventResult xmlns:b="http://schemas.datacontract.org/2004/07/"
                             xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:IsValid>true</b:IsValid>
        <b:StatusCode>00</b:StatusCode>
      </GetStatusEventResult>
    </GetStatusEventResponse>
  </s:Body>
</s:Envelope>`

// TestGetStatusEvent_ParsesResponse confirms GetStatusEvent sends trackId (same shape as
// GetStatus) and parses its own <GetStatusEventResult> element.
func TestGetStatusEvent_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getStatusEventResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetStatusEvent("some-track-id")
	if err != nil {
		t.Fatalf("GetStatusEvent: %v", err)
	}
	if !got.IsValid {
		t.Errorf("IsValid = %v, want true", got.IsValid)
	}
	if !strings.Contains(string(gotBody), "<wcf:trackId>some-track-id</wcf:trackId>") {
		t.Errorf("request body missing trackId: %s", gotBody)
	}
}

const getAcquirerResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetAcquirerResponse xmlns="http://wcf.dian.colombia">
      <GetAcquirerResult>
        <Message>OK</Message>
        <StatusCode>0</StatusCode>
        <ReceiverName>Cliente De Prueba</ReceiverName>
        <ReceiverEmail>cliente@prueba.test</ReceiverEmail>
      </GetAcquirerResult>
    </GetAcquirerResponse>
  </s:Body>
</s:Envelope>`

// TestGetAcquirer_ParsesResponse confirms that GetAcquirer parses a real DIAN response — without
// calling the real server, via an httptest.Server that returns the envelope verbatim.
func TestGetAcquirer_ParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body) // drain the signed request, not inspected here
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getAcquirerResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetAcquirer("31", "900373076")
	if err != nil {
		t.Fatalf("GetAcquirer: %v", err)
	}
	if got.ReceiverName != "Cliente De Prueba" {
		t.Errorf("ReceiverName = %q, want %q", got.ReceiverName, "Cliente De Prueba")
	}
	if got.ReceiverEmail != "cliente@prueba.test" {
		t.Errorf("ReceiverEmail = %q, want %q", got.ReceiverEmail, "cliente@prueba.test")
	}
	if got.StatusCode != "0" {
		t.Errorf("StatusCode = %q, want %q", got.StatusCode, "0")
	}
}

// TestGetAcquirer_NotFoundIsNotAnError confirms the normal/expected case for most ID numbers:
// the DIAN responds without an ErrorMessage/Fault but with empty fields — this must not be
// treated as a Go error, only as "no record found" (non-blocking, see the doc comment on
// Client.GetAcquirer).
func TestGetAcquirer_NotFoundIsNotAnError(t *testing.T) {
	emptyXML := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetAcquirerResponse xmlns="http://wcf.dian.colombia">
      <GetAcquirerResult></GetAcquirerResult>
    </GetAcquirerResponse>
  </s:Body>
</s:Envelope>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(emptyXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetAcquirer("13", "1122334455")
	if err != nil {
		t.Fatalf("GetAcquirer should not fail when there is no record: %v", err)
	}
	if got.ReceiverName != "" || got.ReceiverEmail != "" {
		t.Errorf("expected empty fields, got %+v", got)
	}
}

const getStatusResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetStatusResponse xmlns="http://wcf.dian.colombia">
      <GetStatusResult xmlns:b="http://schemas.datacontract.org/2004/07/"
                        xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:IsValid>true</b:IsValid>
        <b:StatusCode>00</b:StatusCode>
        <b:StatusDescription>La Factura electrónica ha sido autorizada</b:StatusDescription>
      </GetStatusResult>
    </GetStatusResponse>
  </s:Body>
</s:Envelope>`

// TestGetStatus_ParsesResponse confirms GetStatus sends trackId (same request shape as
// GetStatusEvent) and parses its own <GetStatusResult> element — the query counterpart of
// SendBillSync, which this package had never exercised with a test of its own.
func TestGetStatus_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getStatusResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetStatus("some-track-id")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if !got.IsValid {
		t.Errorf("IsValid = %v, want true", got.IsValid)
	}
	if got.StatusCode != "00" {
		t.Errorf("StatusCode = %q, want %q", got.StatusCode, "00")
	}
	if !strings.Contains(string(gotBody), "<wcf:trackId>some-track-id</wcf:trackId>") {
		t.Errorf("request body missing trackId: %s", gotBody)
	}
}

const sendNominaSyncResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <SendNominaSyncResponse xmlns="http://wcf.dian.colombia">
      <SendNominaSyncResult xmlns:b="http://schemas.datacontract.org/2004/07/"
                             xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:IsValid>true</b:IsValid>
        <b:StatusCode>00</b:StatusCode>
        <b:StatusDescription>La Nómina Electrónica ha sido autorizada</b:StatusDescription>
      </SendNominaSyncResult>
    </SendNominaSyncResponse>
  </s:Body>
</s:Envelope>`

// TestSendNominaSync_ParsesResponse confirms SendNominaSync sends only contentFile (no fileName,
// no testSetId — unlike SendNominaSyncTestSet) and parses the shared <SendNominaSyncResult>
// element both functions use.
func TestSendNominaSync_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(sendNominaSyncResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.SendNominaSync([]byte("<NominaIndividual/>"))
	if err != nil {
		t.Fatalf("SendNominaSync: %v", err)
	}
	if !got.IsValid {
		t.Errorf("IsValid = %v, want true", got.IsValid)
	}
	if got.StatusCode != "00" {
		t.Errorf("StatusCode = %q, want %q", got.StatusCode, "00")
	}
	if !strings.Contains(string(gotBody), "<wcf:SendNominaSync>") {
		t.Errorf("request body missing wcf:SendNominaSync element: %s", gotBody)
	}
	if strings.Contains(string(gotBody), "<wcf:testSetId>") {
		t.Errorf("request body should not contain testSetId (that's SendNominaSyncTestSet's job): %s", gotBody)
	}
}

// TestGetAcquirer_HTTP404WithValidBodyIsNotAnError reproduces EXACTLY a real response confirmed
// against the DIAN certification environment (2026-06-29): when the acquirer does not exist,
// the DIAN responds with HTTP status 404 but a perfectly valid SOAP body — StatusCode "404"
// INSIDE the body, Message "El adquirente No existe en la base de datos" ("The acquirer does
// not exist in the database"), with no soap:Fault. Before this test, client.call() rejected
// any status != 200 without inspecting the body, so this normal/expected case was being
// reported as a Go error.
func TestGetAcquirer_HTTP404WithValidBodyIsNotAnError(t *testing.T) {
	realDianNotFoundXML := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://www.w3.org/2005/08/addressing">
  <s:Body>
    <GetAcquirerResponse xmlns="http://wcf.dian.colombia">
      <GetAcquirerResult xmlns:b="http://schemas.datacontract.org/2004/07/Gosocket.Dian.Services.Utils.Common" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:Message>El adquirente No existe en la base de datos</b:Message>
        <b:ReceiverEmail i:nil="true"/>
        <b:ReceiverName i:nil="true"/>
        <b:StatusCode>404</b:StatusCode>
      </GetAcquirerResult>
    </GetAcquirerResponse>
  </s:Body>
</s:Envelope>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(realDianNotFoundXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetAcquirer("13", "6382356")
	if err != nil {
		t.Fatalf("GetAcquirer should not fail on HTTP 404 when the body is valid: %v", err)
	}
	if got.StatusCode != "404" {
		t.Errorf("StatusCode = %q, want %q", got.StatusCode, "404")
	}
	if got.Message != "El adquirente No existe en la base de datos" {
		t.Errorf("Message = %q, does not match the real response", got.Message)
	}
	if got.ReceiverName != "" || got.ReceiverEmail != "" {
		t.Errorf("expected empty fields, got %+v", got)
	}
}

const getXmlByDocumentKeyResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetXmlByDocumentKeyResponse xmlns="http://wcf.dian.colombia">
      <GetXmlByDocumentKeyResult xmlns:b="http://schemas.datacontract.org/2004/07/EventResponse"
                                  xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:Code>00</b:Code>
        <b:Message>OK</b:Message>
        <b:ValidationDate>2026-06-29T10:00:00</b:ValidationDate>
        <b:XmlBytesBase64>PEludm9pY2UvPg==</b:XmlBytesBase64>
      </GetXmlByDocumentKeyResult>
    </GetXmlByDocumentKeyResponse>
  </s:Body>
</s:Envelope>`

// TestGetXmlByDocumentKey_ParsesResponse confirms GetXmlByDocumentKey sends trackId (same
// request shape as GetStatus) and parses its own <GetXmlByDocumentKeyResult> element.
// XmlBytesBase64 arrives as the raw base64 text, not decoded — encoding/xml only copies
// character data verbatim into a []byte field, it never base64-decodes it (verified directly:
// there is no such behavior in the standard library, despite what an earlier, incorrect
// comment in package dian used to claim about DianResponse's equivalent field).
func TestGetXmlByDocumentKey_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getXmlByDocumentKeyResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetXmlByDocumentKey("some-track-id")
	if err != nil {
		t.Fatalf("GetXmlByDocumentKey: %v", err)
	}
	if got.Code != "00" {
		t.Errorf("Code = %q, want %q", got.Code, "00")
	}
	if want := "PEludm9pY2UvPg=="; string(got.XmlBytesBase64) != want {
		t.Errorf("XmlBytesBase64 = %q, want %q (raw base64 text, not decoded)", got.XmlBytesBase64, want)
	}
	if !strings.Contains(string(gotBody), "<wcf:trackId>some-track-id</wcf:trackId>") {
		t.Errorf("request body missing trackId: %s", gotBody)
	}
}

const getReferenceNotesResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetReferenceNotesResponse xmlns="http://wcf.dian.colombia">
      <GetReferenceNotesResult xmlns:b="http://schemas.datacontract.org/2004/07/DianResponse"
                                xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:IsValid>true</b:IsValid>
        <b:StatusCode>00</b:StatusCode>
      </GetReferenceNotesResult>
    </GetReferenceNotesResponse>
  </s:Body>
</s:Envelope>`

// TestGetReferenceNotes_ParsesResponse confirms GetReferenceNotes sends trackId and parses its
// own <GetReferenceNotesResult> element into the shared DianResponse type — matching the real
// WSDL's schema exactly, however unusual that reuse is for a query operation (see the doc
// comment on GetReferenceNotes).
func TestGetReferenceNotes_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getReferenceNotesResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetReferenceNotes("some-track-id")
	if err != nil {
		t.Fatalf("GetReferenceNotes: %v", err)
	}
	if !got.IsValid {
		t.Errorf("IsValid = %v, want true", got.IsValid)
	}
	if !strings.Contains(string(gotBody), "<wcf:GetReferenceNotes>") {
		t.Errorf("request body missing wcf:GetReferenceNotes element: %s", gotBody)
	}
}

const getExchangeEmailsResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetExchangeEmailsResponse xmlns="http://wcf.dian.colombia">
      <GetExchangeEmailsResult xmlns:b="http://schemas.datacontract.org/2004/07/ExchangeEmailResponse"
                                xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:CsvBase64Bytes>bmFtZSxlbWFpbA==</b:CsvBase64Bytes>
        <b:Message>OK</b:Message>
        <b:StatusCode>00</b:StatusCode>
        <b:Success>true</b:Success>
      </GetExchangeEmailsResult>
    </GetExchangeEmailsResponse>
  </s:Body>
</s:Envelope>`

// TestGetExchangeEmails_ParsesResponse confirms GetExchangeEmails sends no parameters (an empty
// self-closing <wcf:GetExchangeEmails/> element, per the WSDL's empty request schema) and
// parses its own <GetExchangeEmailsResult> element. CsvBase64Bytes arrives as raw base64 text —
// see XmlBytesBase64's test for why it isn't decoded automatically.
func TestGetExchangeEmails_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getExchangeEmailsResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetExchangeEmails()
	if err != nil {
		t.Fatalf("GetExchangeEmails: %v", err)
	}
	if !got.Success {
		t.Errorf("Success = %v, want true", got.Success)
	}
	if want := "bmFtZSxlbWFpbA=="; string(got.CsvBase64Bytes) != want {
		t.Errorf("CsvBase64Bytes = %q, want %q (raw base64 text, not decoded)", got.CsvBase64Bytes, want)
	}
	if !strings.Contains(string(gotBody), "<wcf:GetExchangeEmails/>") {
		t.Errorf("request body missing self-closing wcf:GetExchangeEmails element: %s", gotBody)
	}
}

const getDocumentInfoResponseXML = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <GetDocumentInfoResponse xmlns="http://wcf.dian.colombia">
      <GetDocumentInfoResult xmlns:b="http://schemas.datacontract.org/2004/07/DocumentInfoResponse"
                              xmlns:c="http://schemas.datacontract.org/2004/07/Documento"
                              xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <b:StatusCode>00</b:StatusCode>
        <b:StatusDescription>Proceso Exitoso</b:StatusDescription>
        <b:DocumentInfo>
          <c:Documento>
            <c:DocumentCode>SETP990000001</c:DocumentCode>
            <c:DocumentTypeId>01</c:DocumentTypeId>
            <c:DocumentTypeName>Factura Electrónica de Venta</c:DocumentTypeName>
            <c:Emisor>
              <Nombre xmlns="http://schemas.datacontract.org/2004/07/Entidad">MI EMPRESA S.A.S.</Nombre>
              <NumeroDoc xmlns="http://schemas.datacontract.org/2004/07/Entidad">900123456</NumeroDoc>
            </c:Emisor>
            <c:TotalEImpuestos>
              <Iva xmlns="http://schemas.datacontract.org/2004/07/TotalEImpuestos">239495.80</Iva>
              <Total xmlns="http://schemas.datacontract.org/2004/07/TotalEImpuestos">1500000.00</Total>
            </c:TotalEImpuestos>
            <c:UUID>18015e1f4f6b1eb55cf6d5eaa1f752bed3b0402e0cf11eb515c1ce5ccbe9bca120cd4776ee3b1e5c281e0fd2711d40d1</c:UUID>
          </c:Documento>
        </b:DocumentInfo>
      </GetDocumentInfoResult>
    </GetDocumentInfoResponse>
  </s:Body>
</s:Envelope>`

// TestGetDocumentInfo_ParsesResponse confirms GetDocumentInfo sends uuid (not trackId, unlike
// most other queries) and parses the nested DocumentInfoResponse/Documento/Emisor/
// TotalEImpuestos structure — proving at least one full level of nesting decodes correctly,
// without exhaustively exercising every leaf type declared in the WSDL.
func TestGetDocumentInfo_ParsesResponse(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(getDocumentInfoResponseXML))
	}))
	defer server.Close()

	cert, key := generateTestCert(t)
	c := &Client{BaseURL: server.URL, Cert: cert, Key: key, HTTPClient: server.Client()}

	got, err := c.GetDocumentInfo("18015e1f4f6b1eb55cf6d5eaa1f752bed3b0402e0cf11eb515c1ce5ccbe9bca120cd4776ee3b1e5c281e0fd2711d40d1")
	if err != nil {
		t.Fatalf("GetDocumentInfo: %v", err)
	}
	if got.StatusCode != "00" {
		t.Errorf("StatusCode = %q, want %q", got.StatusCode, "00")
	}
	if len(got.DocumentInfo) != 1 {
		t.Fatalf("len(DocumentInfo) = %d, want 1", len(got.DocumentInfo))
	}
	doc := got.DocumentInfo[0]
	if doc.DocumentCode != "SETP990000001" {
		t.Errorf("DocumentCode = %q, want %q", doc.DocumentCode, "SETP990000001")
	}
	if doc.Emisor.Nombre != "MI EMPRESA S.A.S." {
		t.Errorf("Emisor.Nombre = %q, want %q", doc.Emisor.Nombre, "MI EMPRESA S.A.S.")
	}
	if doc.TotalEImpuestos.Total != 1500000.00 {
		t.Errorf("TotalEImpuestos.Total = %v, want %v", doc.TotalEImpuestos.Total, 1500000.00)
	}
	if !strings.Contains(string(gotBody), "<wcf:uuid>") {
		t.Errorf("request body missing wcf:uuid element (GetDocumentInfo uses uuid, not trackId): %s", gotBody)
	}
}
