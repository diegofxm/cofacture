// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package soap

import (
	"encoding/base64"

	"github.com/beevik/etree"
)

// SendTestSetAsync sends a ZIP for the certification test set. Returns the
// UploadDocumentResponse — if ZipKey comes back empty, check ErrorMessageList (DIAN rejected
// the ZIP during initial validation, before queueing it). The actual validation result is
// queried afterward with GetStatusZip(ZipKey).
func (c *Client) SendTestSetAsync(fileName string, content []byte, testSetID string) (*UploadDocumentResponse, error) {
	var result struct {
		Result UploadDocumentResponse `xml:"SendTestSetAsyncResult"`
	}
	err := c.call("SendTestSetAsync", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendTestSetAsync")
		el.CreateElement("wcf:fileName").SetText(fileName)
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString(content))
		el.CreateElement("wcf:testSetId").SetText(testSetID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// SendBillSync sends a ZIP with a single UBL document and returns the validation result
// immediately (a synchronous process, Technical Annex 1.9 section 7.10) — unlike
// SendTestSetAsync/SendBillAsync, there is no ZipKey or later query step; the DianResponse
// itself already carries StatusCode/IsValid.
func (c *Client) SendBillSync(fileName string, content []byte) (*DianResponse, error) {
	var result struct {
		Result DianResponse `xml:"SendBillSyncResult"`
	}
	err := c.call("SendBillSync", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendBillSync")
		el.CreateElement("wcf:fileName").SetText(fileName)
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString(content))
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// SendBillAsync sends a ZIP with one or more UBL documents asynchronously (Technical Annex 1.9
// section 7.8) — same pattern as SendTestSetAsync (returns a ZipKey, the actual result is
// queried later with GetStatusZip), but without testSetId: it's for normal submissions, not
// the certification test set.
func (c *Client) SendBillAsync(fileName string, content []byte) (*UploadDocumentResponse, error) {
	var result struct {
		Result UploadDocumentResponse `xml:"SendBillAsyncResult"`
	}
	err := c.call("SendBillAsync", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendBillAsync")
		el.CreateElement("wcf:fileName").SetText(fileName)
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString(content))
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// SendBillAttachmentAsync sends a ZIP asynchronously, same request/response shape as
// SendBillAsync (fileName + contentFile -> UploadDocumentResponse, poll with GetStatusZip).
// Confirmed present in the real certification-environment WSDL
// (WcfDianCustomerServices.svc?singleWsdl) as its own operation, distinct from SendBillAsync.
func (c *Client) SendBillAttachmentAsync(fileName string, content []byte) (*UploadDocumentResponse, error) {
	var result struct {
		Result UploadDocumentResponse `xml:"SendBillAttachmentAsyncResult"`
	}
	err := c.call("SendBillAttachmentAsync", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendBillAttachmentAsync")
		el.CreateElement("wcf:fileName").SetText(fileName)
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString(content))
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetStatusZip queries the validation result of a ZIP sent via SendBillAsync or
// SendTestSetAsync. A ZIP can contain several documents, hence the slice return.
func (c *Client) GetStatusZip(trackID string) ([]DianResponse, error) {
	var result struct {
		Result struct {
			Items []DianResponse `xml:"DianResponse"`
		} `xml:"GetStatusZipResult"`
	}
	err := c.call("GetStatusZip", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetStatusZip")
		el.CreateElement("wcf:trackId").SetText(trackID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return result.Result.Items, nil
}

// GetStatus queries the validation result of a single document sent via SendBillSync.
func (c *Client) GetStatus(trackID string) (*DianResponse, error) {
	var result struct {
		Result DianResponse `xml:"GetStatusResult"`
	}
	err := c.call("GetStatus", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetStatus")
		el.CreateElement("wcf:trackId").SetText(trackID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetNumberingRange queries the numbering ranges DIAN has authorized for a given issuer +
// software pair. accountCode and accountCodeT are the issuer's NIT (identical in direct
// integration); softwareCode is the UUID of the registered software.
// Returns every active and historical range DIAN has for that issuer/software pair, each with
// its resolution, prefix, from/to, validity dates and the TechnicalKey needed to compute the
// CUFE.
func (c *Client) GetNumberingRange(accountCode, accountCodeT, softwareCode string) (*NumberRangeResponseList, error) {
	var result struct {
		Result NumberRangeResponseList `xml:"GetNumberingRangeResult"`
	}
	err := c.call("GetNumberingRange", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetNumberingRange")
		el.CreateElement("wcf:accountCode").SetText(accountCode)
		el.CreateElement("wcf:accountCodeT").SetText(accountCodeT)
		el.CreateElement("wcf:softwareCode").SetText(softwareCode)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// SendNominaSync sends a ZIP with a signed NominaIndividual (or NominaIndividualDeAjuste) and
// returns the validation result synchronously (Payroll Technical Annex, section 9.7). Unlike
// SendBillSync, it only takes contentFile (no fileName).
func (c *Client) SendNominaSync(content []byte) (*DianResponse, error) {
	var result struct {
		Result DianResponse `xml:"SendNominaSyncResult"`
	}
	err := c.call("SendNominaSync", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendNominaSync")
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString(content))
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// SendNominaSyncTestSet sends a payroll ZIP to the certification test set. Combines
// SendNominaSync's logic with the testSetId required for certification.
func (c *Client) SendNominaSyncTestSet(content []byte, testSetID string) (*DianResponse, error) {
	var result struct {
		Result DianResponse `xml:"SendNominaSyncResult"`
	}
	err := c.call("SendNominaSync", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendNominaSync")
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString(content))
		el.CreateElement("wcf:testSetId").SetText(testSetID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// SendEventUpdateStatus submits a signed ApplicationResponse event (Acuse de Recibo, Reclamo,
// Recibo del Bien, Aceptación Expresa/Tácita — see package event and builder.BuildAcuseRecibo
// and friends) and returns the validation result synchronously, same response shape as
// SendBillSync/SendNominaSync. Confirmed present in the real certification-environment WSDL
// (WcfDianCustomerServices.svc?singleWsdl) as its own operation.
func (c *Client) SendEventUpdateStatus(content []byte) (*DianResponse, error) {
	var result struct {
		Result DianResponse `xml:"SendEventUpdateStatusResult"`
	}
	err := c.call("SendEventUpdateStatus", func(body *etree.Element) {
		el := body.CreateElement("wcf:SendEventUpdateStatus")
		el.CreateElement("wcf:contentFile").SetText(base64.StdEncoding.EncodeToString(content))
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetStatusEvent queries the validation result of an event sent via SendEventUpdateStatus, same
// pattern as GetStatus for SendBillSync.
func (c *Client) GetStatusEvent(trackID string) (*DianResponse, error) {
	var result struct {
		Result DianResponse `xml:"GetStatusEventResult"`
	}
	err := c.call("GetStatusEvent", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetStatusEvent")
		el.CreateElement("wcf:trackId").SetText(trackID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetAcquirer queries DIAN's exchange/notification registry for a third party — see
// AcquirerResponse. An optional aid when capturing a NIT, never blocking: an empty result (no
// ReceiverName/ReceiverEmail) is normal and expected for most identification numbers.
func (c *Client) GetAcquirer(identificationType, identificationNumber string) (*AcquirerResponse, error) {
	var result struct {
		Result AcquirerResponse `xml:"GetAcquirerResult"`
	}
	err := c.call("GetAcquirer", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetAcquirer")
		el.CreateElement("wcf:identificationType").SetText(identificationType)
		el.CreateElement("wcf:identificationNumber").SetText(identificationNumber)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetXmlByDocumentKey retrieves the original signed XML (base64-encoded, in
// XMLByDocumentKeyResponse.XmlBytesBase64) of a document previously submitted via
// SendBillSync/SendBillAsync/SendTestSetAsync, given the TrackId/document key returned at
// submission time.
func (c *Client) GetXmlByDocumentKey(trackID string) (*XMLByDocumentKeyResponse, error) {
	var result struct {
		Result XMLByDocumentKeyResponse `xml:"GetXmlByDocumentKeyResult"`
	}
	err := c.call("GetXmlByDocumentKey", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetXmlByDocumentKey")
		el.CreateElement("wcf:trackId").SetText(trackID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetReferenceNotes queries which Credit/Debit Notes reference the document identified by
// trackID — a reverse lookup. Its result is shaped as DianResponse in the real WSDL (the same
// type SendBillSync/GetStatus use), which is unusual for a query operation; this has not been
// exercised against DIAN's certification environment, so which of DianResponse's fields
// actually carry data here (as opposed to being left empty) is unconfirmed.
func (c *Client) GetReferenceNotes(trackID string) (*DianResponse, error) {
	var result struct {
		Result DianResponse `xml:"GetReferenceNotesResult"`
	}
	err := c.call("GetReferenceNotes", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetReferenceNotes")
		el.CreateElement("wcf:trackId").SetText(trackID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetDocumentInfo queries full metadata for a document identified by UUID (parties, taxes,
// referencing notes/events, validations) without downloading its XML — a heavier query than
// GetStatus, lighter than GetXmlByDocumentKey. See DocumentInfoResponse's doc comment for the
// same not-yet-confirmed-against-real-DIAN caveat as GetReferenceNotes.
func (c *Client) GetDocumentInfo(documentUUID string) (*DocumentInfoResponse, error) {
	var result struct {
		Result DocumentInfoResponse `xml:"GetDocumentInfoResult"`
	}
	err := c.call("GetDocumentInfo", func(body *etree.Element) {
		el := body.CreateElement("wcf:GetDocumentInfo")
		el.CreateElement("wcf:uuid").SetText(documentUUID)
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

// GetExchangeEmails queries the email addresses configured to exchange electronic documents
// with DIAN's registry, as a base64-encoded CSV (ExchangeEmailResponse.CsvBase64Bytes). Takes
// no parameters — confirmed in the real WSDL's schema (an empty sequence).
func (c *Client) GetExchangeEmails() (*ExchangeEmailResponse, error) {
	var result struct {
		Result ExchangeEmailResponse `xml:"GetExchangeEmailsResult"`
	}
	err := c.call("GetExchangeEmails", func(body *etree.Element) {
		body.CreateElement("wcf:GetExchangeEmails")
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}
