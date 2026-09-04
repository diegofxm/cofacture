// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
)

// appendPaymentMean adds cac:PaymentMeans — shared by Invoice, CreditNote and DebitNote (same
// element, same content in all three).
func appendPaymentMean(parent *etree.Element, pm domain.PaymentMean) {
	el := parent.CreateElement("cac:PaymentMeans")
	el.CreateElement("cbc:ID").SetText(pm.Code)
	el.CreateElement("cbc:PaymentMeansCode").SetText(pm.PaymentMethodCode)
	if pm.Code == "2" {
		el.CreateElement("cbc:PaymentDueDate").SetText(pm.DueDate)
	}
	if pm.PaymentReference != "" {
		el.CreateElement("cbc:PaymentID").SetText(pm.PaymentReference)
	}
}
