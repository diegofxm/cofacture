// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
)

// appendDiscrepancyResponse adds cac:DiscrepancyResponse — the note's reason. Only called when
// the note actually carries one (operation "20" for Credit Note, "30" for Debit Note).
func appendDiscrepancyResponse(parent *etree.Element, dr domain.DiscrepancyResponse) {
	el := parent.CreateElement("cac:DiscrepancyResponse")
	el.CreateElement("cbc:ReferenceID").SetText(dr.ReferenceID)
	el.CreateElement("cbc:ResponseCode").SetText(dr.ResponseCode)
	el.CreateElement("cbc:Description").SetText(dr.Description)
}

// appendBillingReference adds cac:BillingReference — the reference to the invoice this note
// corrects. node is "InvoiceDocumentReference" for both Credit Note and Debit Note. schemeName
// comes from br.HashType because the referenced document is not always a regular Invoice
// (CUFE-SHA384) — a Credit/Debit Note adjusting a Documento Equivalente Electrónico (DocumentType
// "93"/"94") references a CUDE-SHA384 document instead.
func appendBillingReference(parent *etree.Element, node string, br domain.BillingReference) {
	ref := parent.CreateElement("cac:BillingReference").CreateElement("cac:" + node)
	ref.CreateElement("cbc:ID").SetText(br.Prefix + br.Number)
	uuid := ref.CreateElement("cbc:UUID")
	uuid.CreateAttr("schemeName", br.HashType)
	uuid.SetText(br.CUFE)
	ref.CreateElement("cbc:IssueDate").SetText(br.IssueDate)
}
