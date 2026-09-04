// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"os"
	"testing"

	"github.com/diegofxm/cofacture/domain"
	"github.com/diegofxm/cofacture/event"
)

func sampleEvent() domain.Event {
	return domain.Event{
		EnvironmentCode: "2",
		ID:              "1",
		IssueDate:       "2024-02-05",
		IssueTime:       "09:00:00-05:00",

		DocumentReference: domain.EventDocumentReference{
			Prefix:           "SETP",
			Number:           "990068706",
			CUFE:             "853657dcf2841c55c04338b24cc4db9dfbf87042f1ce1798e53f7b1f0502d00df9bd3f371dea47b02766424976d60ba2",
			HashType:         "CUFE-SHA384",
			DocumentTypeCode: "01",
		},

		Sender: domain.EventParty{
			Name:           "Consumidor Final",
			Identification: domain.Identification{Number: "222222222222", TypeCode: "13"},
			TaxSchemeCode:  "ZZ",
			TaxSchemeName:  "No aplica",
		},
		Receiver: domain.EventParty{
			Name:           "MI EMPRESA S.A.S.",
			Identification: domain.Identification{Number: "900123456", TypeCode: "31", VerificationCode: "3"},
			TaxSchemeCode:  "01",
			TaxSchemeName:  "IVA",
		},

		SoftwareProvider: domain.SoftwareProvider{
			ProviderIdentification: domain.Identification{Number: "900123456", TypeCode: "31", VerificationCode: "3"},
			SoftwareID:             "12345678-1234-1234-1234-123456789012",
		},

		CUDE:                 "0d91ba25b01f5e7dbda870a11b274501d3a62a73e91932c473c86c93f12a142a2ac45876efcde3e679024a01c0be41f9",
		SoftwareSecurityCode: "abc123",
		QRURL:                "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=853657dcf2841c55c04338b24cc4db9dfbf87042f1ce1798e53f7b1f0502d00df9bd3f371dea47b02766424976d60ba2",
	}
}

func goldenTest(t *testing.T, path string, build func() (interface{ WriteToString() (string, error) }, error)) {
	t.Helper()
	docAny, err := build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	type indenter interface{ Indent(int) }
	if id, ok := docAny.(indenter); ok {
		id.Indent(2)
	}
	got, err := docAny.WriteToString()
	if err != nil {
		t.Fatalf("WriteToString: %v", err)
	}

	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Skip("golden file regenerated, review by hand before trusting it")
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (missing -update run?): %v", err)
	}
	if got != string(want) {
		t.Errorf("generated XML does not match %s\n--- got ---\n%s", path, got)
	}
}

func TestBuildAcuseRecibo_Golden(t *testing.T) {
	ev := sampleEvent()
	ev.ReceiverPerson = &domain.EventReceiverPerson{
		Identification: domain.Identification{Number: "1234567890", TypeCode: "13"},
		FirstName:      "Juan",
		FamilyName:     "Pérez",
		JobTitle:       "Gerente de Compras",
	}
	goldenTest(t, "testdata/acuse_recibo_golden.xml", func() (interface{ WriteToString() (string, error) }, error) {
		return BuildAcuseRecibo(ev)
	})
}

func TestBuildAcuseRecibo_RequiresReceiverPerson(t *testing.T) {
	if _, err := BuildAcuseRecibo(sampleEvent()); err == nil {
		t.Error("expected an error when ReceiverPerson is nil")
	}
}

func TestBuildRecibidoBien_Golden(t *testing.T) {
	ev := sampleEvent()
	goldenTest(t, "testdata/recibido_bien_golden.xml", func() (interface{ WriteToString() (string, error) }, error) {
		return BuildRecibidoBien(ev)
	})
}

func TestBuildAceptacionExpresa_Golden(t *testing.T) {
	ev := sampleEvent()
	goldenTest(t, "testdata/aceptacion_expresa_golden.xml", func() (interface{ WriteToString() (string, error) }, error) {
		return BuildAceptacionExpresa(ev)
	})
}

func TestBuildAceptacionTacita_Golden(t *testing.T) {
	ev := sampleEvent()
	// Aceptación Tácita is issuer-generated: DIAN is the recipient of the event, and the issuer
	// (Receiver in the referenced invoice) is who sends it — roles are swapped vs. the other
	// four events, which is why Sender/Receiver are reversed here relative to sampleEvent().
	ev.Sender, ev.Receiver = ev.Receiver, ev.Sender
	ev.Note = event.TacitAcceptanceNote("1", ev.CUDE, "Consumidor Final", "222222222222")
	goldenTest(t, "testdata/aceptacion_tacita_golden.xml", func() (interface{ WriteToString() (string, error) }, error) {
		return BuildAceptacionTacita(ev)
	})
}

func TestBuildReclamo_Golden(t *testing.T) {
	r := domain.Reclamo{
		Event:           sampleEvent(),
		RejectionListID: "2",
		RejectionName:   "Reclamo",
	}
	goldenTest(t, "testdata/reclamo_golden.xml", func() (interface{ WriteToString() (string, error) }, error) {
		return BuildReclamo(r)
	})
}
