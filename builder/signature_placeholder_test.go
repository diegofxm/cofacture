// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import "testing"

// These tests confirm that SignaturePlaceholder behaves the same way for all three
// document types — all three reuse appendUBLExtensions, but this provides an early
// warning if any builder's structure ever diverges.

func TestSignaturePlaceholder_Invoice(t *testing.T) {
	doc, err := BuildInvoice(sampleInvoice())
	if err != nil {
		t.Fatalf("BuildInvoice: %v", err)
	}
	if _, err := SignaturePlaceholder(doc); err != nil {
		t.Errorf("SignaturePlaceholder: %v", err)
	}
}

func TestSignaturePlaceholder_CreditNote(t *testing.T) {
	doc, err := BuildCreditNote(sampleCreditNote())
	if err != nil {
		t.Fatalf("BuildCreditNote: %v", err)
	}
	if _, err := SignaturePlaceholder(doc); err != nil {
		t.Errorf("SignaturePlaceholder: %v", err)
	}
}

func TestSignaturePlaceholder_DebitNote(t *testing.T) {
	doc, err := BuildDebitNote(sampleDebitNote())
	if err != nil {
		t.Fatalf("BuildDebitNote: %v", err)
	}
	if _, err := SignaturePlaceholder(doc); err != nil {
		t.Errorf("SignaturePlaceholder: %v", err)
	}
}
