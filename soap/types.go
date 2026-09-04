// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package soap

// Types that mirror the real schemas downloaded from the certification-environment WSDL
// (https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?xsd=xsd2, xsd3, xsd5, xsd6, xsd9,
// xsd11-xsd21) — they are not a translation of the technical annex, they are the service's
// actual contract as exposed.
//
// The xml tags do not qualify a namespace: encoding/xml matches by local name when the tag
// specifies none, and there is no possible name collision across namespaces here.

type stringArray struct {
	Items []string `xml:"string"`
}

// DianResponse is the result of validating a document (SendBillSyncResult,
// GetStatusResult, each element of GetStatusZipResult).
type DianResponse struct {
	ErrorMessage      *stringArray `xml:"ErrorMessage"`
	IsValid           bool         `xml:"IsValid"`
	StatusCode        string       `xml:"StatusCode"`
	StatusDescription string       `xml:"StatusDescription"`
	StatusMessage     string       `xml:"StatusMessage"`
	XmlBase64Bytes    []byte       `xml:"XmlBase64Bytes"`
	XmlBytes          []byte       `xml:"XmlBytes"`
	XmlDocumentKey    string       `xml:"XmlDocumentKey"`
	XmlFileName       string       `xml:"XmlFileName"`
}

// XMLParamsResponseTrackId is an initial ZIP validation error (it never made it into the
// validation queue).
type XMLParamsResponseTrackId struct {
	DocumentKey      string `xml:"DocumentKey"`
	ProcessedMessage string `xml:"ProcessedMessage"`
	SenderCode       string `xml:"SenderCode"`
	Success          bool   `xml:"Success"`
	XmlFileName      string `xml:"XmlFileName"`
}

// UploadDocumentResponse is the result of SendBillAsync/SendTestSetAsync/
// SendBillAttachmentAsync: either there are initial errors (ErrorMessageList) or there is a
// ZipKey to query later with GetStatusZip.
type UploadDocumentResponse struct {
	ErrorMessageList *struct {
		Items []XMLParamsResponseTrackId `xml:"XmlParamsResponseTrackId"`
	} `xml:"ErrorMessageList"`
	ZipKey string `xml:"ZipKey"`
}

// NumberRangeResponse is a single numbering range returned by GetNumberingRange (namespace
// http://schemas.datacontract.org/2004/07/NumberRangeResponse in the real WSDL). Dates are
// plain strings exactly as DIAN returns them ("2019-01-19T00:00:00") — the caller converts as
// needed.
type NumberRangeResponse struct {
	ResolutionNumber string `xml:"ResolutionNumber"`
	ResolutionDate   string `xml:"ResolutionDate"`
	Prefix           string `xml:"Prefix"`
	FromNumber       int64  `xml:"FromNumber"`
	ToNumber         int64  `xml:"ToNumber"`
	ValidDateFrom    string `xml:"ValidDateFrom"`
	ValidDateTo      string `xml:"ValidDateTo"`
	TechnicalKey     string `xml:"TechnicalKey"`
}

// NumberRangeResponseList is the full result of GetNumberingRange — a list of ranges
// authorized for the given issuer + software pair (namespace
// http://schemas.datacontract.org/2004/07/NumberRangeResponseList). OperationCode "0" means
// success; any other code indicates a DIAN-side error.
type NumberRangeResponseList struct {
	OperationCode        string                `xml:"OperationCode"`
	OperationDescription string                `xml:"OperationDescription"`
	ResponseList         []NumberRangeResponse `xml:"ResponseList>NumberRangeResponse"`
}

// AcquirerResponse is the result of GetAcquirer (namespace
// Gosocket.Dian.Services.Utils.Common in the real WSDL) — it queries DIAN's exchange/
// notification registry for a given third party, not a full RUT lookup: it only returns
// ReceiverName/ReceiverEmail if that identificationNumber already has a name/email registered
// to receive electronic documents. Empty StatusCode/Message (or an error StatusCode) is the
// normal, expected result for most national IDs — it does not mean the query failed.
type AcquirerResponse struct {
	Message       string `xml:"Message"`
	StatusCode    string `xml:"StatusCode"`
	ReceiverName  string `xml:"ReceiverName"`
	ReceiverEmail string `xml:"ReceiverEmail"`
}

// XMLByDocumentKeyResponse is the result of GetXmlByDocumentKey. DIAN's own WSDL names this
// datacontract type "EventResponse" (http://schemas.datacontract.org/2004/07/EventResponse) —
// a naming collision with this library's own event package that has nothing to do with it;
// this Go type is named for what it actually holds instead of copying that name.
type XMLByDocumentKeyResponse struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
	// XmlBytesBase64 is the raw base64 text as DIAN sends it — encoding/xml does not decode
	// []byte fields from base64 (it only copies character data verbatim), so the caller must
	// call base64.StdEncoding.DecodeString on this before it's usable XML. Same convention as
	// DianResponse.XmlBase64Bytes.
	XmlBytesBase64 []byte `xml:"XmlBytesBase64"`
	ValidationDate string `xml:"ValidationDate"`
}

// ExchangeEmailResponse is the result of GetExchangeEmails — a CSV of the email addresses
// configured to exchange electronic documents with DIAN's registry.
type ExchangeEmailResponse struct {
	// CsvBase64Bytes is the raw base64 text as DIAN sends it — see XMLByDocumentKeyResponse's
	// doc comment on XmlBytesBase64 for why this isn't decoded automatically.
	CsvBase64Bytes []byte `xml:"CsvBase64Bytes"`
	Message        string `xml:"Message"`
	StatusCode     string `xml:"StatusCode"`
	Success        bool   `xml:"Success"`
}

// DocumentInfoResponse is the result of GetDocumentInfo — full metadata (parties, taxes,
// notes, referencing events, validations) for a document identified by UUID (namespace
// http://schemas.datacontract.org/2004/07/DocumentInfoResponse in the real WSDL). This is a
// heavier query than GetStatus and has not been exercised against DIAN's certification
// environment — the shape below mirrors the WSDL's own schema exactly, but which fields DIAN
// actually populates in practice is unconfirmed.
type DocumentInfoResponse struct {
	CompressedDocumentInfo string      `xml:"CompressedDocumentInfo"`
	DocumentInfo           []Documento `xml:"DocumentInfo>Documento"`
	StatusCode             string      `xml:"StatusCode"`
	StatusDescription      string      `xml:"StatusDescription"`
}

// Documento is one entry of DocumentInfoResponse.DocumentInfo.
type Documento struct {
	DocumentCode        string                `xml:"DocumentCode"`
	DocumentDescription string                `xml:"DocumentDescription"`
	DocumentTags        []Nota                `xml:"DocumentTags>Nota"`
	DocumentTypeId      string                `xml:"DocumentTypeId"`
	DocumentTypeName    string                `xml:"DocumentTypeName"`
	Emisor              Entidad               `xml:"Emisor"`
	Estado              []intStringPair       `xml:"Estado>KeyValueOfintstring"`
	Eventos             []Evento              `xml:"Eventos>Evento"`
	LegitimoTenedor     LegitimoTenedor       `xml:"LegitimoTenedor"`
	NumeroDocumento     NumeroDocumento       `xml:"NumeroDocumento"`
	Receptor            Entidad               `xml:"Receptor"`
	Referencias         []ReferenciaDocumento `xml:"Referencias>ReferenciaDocumento"`
	TotalEImpuestos     TotalEImpuestos       `xml:"TotalEImpuestos"`
	UUID                string                `xml:"UUID"`
	ValidacionesDoc     []ValidacionDoc       `xml:"ValidacionesDoc>ValidacionDoc"`
}

// intStringPair is Documento.Estado's dictionary entry (ArrayOfKeyValueOfintstring in the WSDL
// — DIAN represents Estado as an int-keyed dictionary, not a plain string).
type intStringPair struct {
	Key   int    `xml:"Key"`
	Value string `xml:"Value"`
}

// Nota is an entry of Documento.DocumentTags. Despite the generic name, the real WSDL gives it
// full party and correction-concept data — it is not free-text.
type Nota struct {
	ConceptoCorreccion  ConceptoCorreccion `xml:"ConceptoCorreccion"`
	Emisor              Entidad            `xml:"Emisor"`
	LegitimoTenedor     LegitimoTenedor    `xml:"LegitimoTenedor"`
	NombreTipoDocumento string             `xml:"NombreTipoDocumento"`
	NumeroDocumento     NumeroDocumento    `xml:"NumeroDocumento"`
	Receptor            Entidad            `xml:"Receptor"`
	UUID                string             `xml:"UUID"`
	ValidacionesDoc     []ValidacionDoc    `xml:"ValidacionesDoc>ValidacionDoc"`
}

// ConceptoCorreccion is a Credit/Debit Note's correction-concept catalog entry.
type ConceptoCorreccion struct {
	Codigo      string `xml:"Codigo"`
	Descripcion string `xml:"Descripcion"`
	Nombre      string `xml:"Nombre"`
}

// Entidad identifies a party (issuer or receiver) in GetDocumentInfo's response — a much
// smaller shape than domain.Party, since this is DIAN's own summary, not the full UBL party.
type Entidad struct {
	Nombre      string `xml:"Nombre"`
	NumeroDoc   string `xml:"NumeroDoc"`
	Procedencia string `xml:"Procedencia"`
	TipoDoc     string `xml:"TipoDoc"`
}

// LegitimoTenedor is the "legitimate holder" of a title-value document (factura como título
// valor) — populated only after it has been transferred/endorsed.
type LegitimoTenedor struct {
	FechaInscripcionComoTituloValor string `xml:"FechaInscripcionComoTituloValor"`
	Nombre                          string `xml:"Nombre"`
}

// NumeroDocumento carries a document's series/folio and its issuance/signature dates.
type NumeroDocumento struct {
	FechaEmision string `xml:"FechaEmision"`
	FechaFirma   string `xml:"FechaFirma"`
	Folio        string `xml:"Folio"`
	Serie        string `xml:"Serie"`
}

// TotalEImpuestos is a document's IVA and total amount, as DIAN's own summary reports them —
// a bare xs:double from the wire, not this library's int64-cents convention: there is nothing
// to truncate here, this is DIAN's own already-formatted response, not a value we construct.
type TotalEImpuestos struct {
	Iva   float64 `xml:"Iva"`
	Total float64 `xml:"Total"`
}

// ValidacionDoc is one validation entry DIAN ran against a document (or one of its events).
type ValidacionDoc struct {
	IsNotification bool   `xml:"IsNotification"`
	IsValida       bool   `xml:"IsValida"`
	MensajeError   string `xml:"MensajeError"`
	Nombre         string `xml:"Nombre"`
	Status         string `xml:"Status"`
}

// Evento is one lifecycle event recorded against a document (e.g. Acuse de Recibo, Aceptación
// Tácita) as summarized by GetDocumentInfo — not to be confused with this library's own event
// package, which builds those events before submission; this type only describes what DIAN
// already has on file.
type Evento struct {
	Codigo               string                `xml:"Codigo"`
	Descripcion          string                `xml:"Descripcion"`
	Emisor               Entidad               `xml:"Emisor"`
	NumeroDocumento      NumeroDocumento       `xml:"NumeroDocumento"`
	Receptor             Entidad               `xml:"Receptor"`
	ReferenciasDocumento []ReferenciaDocumento `xml:"ReferenciasDocumento>ReferenciaDocumento"`
	UUID                 string                `xml:"UUID"`
	ValidacionesDoc      []ValidacionDoc       `xml:"ValidacionesDoc>ValidacionDoc"`
}

// ReferenciaDocumento is a reference to another document (e.g. the invoice a Credit Note
// corrects), as summarized by GetDocumentInfo/Evento.
type ReferenciaDocumento struct {
	Descripcion      string  `xml:"Descripcion"`
	DocumentTypeId   string  `xml:"DocumentTypeId"`
	DocumentTypeName string  `xml:"DocumentTypeName"`
	Emisor           Entidad `xml:"Emisor"`
	Fecha            string  `xml:"Fecha"`
	Receptor         Entidad `xml:"Receptor"`
	UUID             string  `xml:"UUID"`
}
