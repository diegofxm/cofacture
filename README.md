# cofacture

[![Go Reference](https://pkg.go.dev/badge/github.com/diegofxm/cofacture.svg)](https://pkg.go.dev/github.com/diegofxm/cofacture)
[![Go Version](https://img.shields.io/github/go-mod/go-version/diegofxm/cofacture)](go.mod)
[![License](https://img.shields.io/badge/license-AGPL--3.0-blue)](LICENSE)

A Go library for Colombian electronic invoicing (DIAN): builds UBL 2.1 documents, computes CUFE/CUDE/CUDS/CUNE, signs with XAdES-EPES, and talks to DIAN's `WcfDianCustomerServices` SOAP web service.

`cofacture` is a **library, not a platform** — no database, no HTTP server, no opinion about how you store your data. You feed it plain Go structs; it hands back signed XML, ready to send. You wire the pipeline (build → hash → sign → zip → send) yourself.

---

## What it does

- Builds every DIAN electronic document type defined over UBL 2.1: Electronic Sales Invoice, Credit Note, Debit Note, Support Document, Adjustment Note to the Support Document, Attached Document, and Documento Equivalente Electrónico (POS ticket and its other sub-types).
- Builds Individual Electronic Payroll and its adjustment/cancellation variants (`NominaIndividual`, not UBL).
- Builds the five RADIAN events: Acuse de Recibo, Reclamo, Recibo del Bien, Aceptación Expresa, Aceptación Tácita.
- Computes CUFE, CUDE, CUDS, and CUNE per the DIAN Technical Annexes.
- Signs documents with XAdES-EPES (C14N 1.0, RSA-SHA256), loading certificates from PEM or PKCS#12.
- Packages signed XML using the ZIP format and file-naming convention DIAN's receiving service requires.
- Implements all 16 operations of the `WcfDianCustomerServices` SOAP 1.2 + WS-Security web service (habilitación and producción): `SendBillSync`, `SendBillAsync`, `SendBillAttachmentAsync`, `SendTestSetAsync`, `GetStatus`, `GetStatusZip`, `GetNumberingRange`, `SendNominaSync`, `SendNominaSyncTestSet`, `SendEventUpdateStatus`, `GetStatusEvent`, `GetAcquirer`, `GetXmlByDocumentKey`, `GetReferenceNotes`, `GetDocumentInfo`, `GetExchangeEmails`.
- Parses DIAN's validation responses into a structured result (rejections vs. notices, embedded `ApplicationResponse`, etc.).

## Design boundaries

cofacture builds and sends documents; it does not decide what goes in them:

- **No catalog validation.** Tax types, unit codes, city/DANE codes, payment methods, and similar codes are placed exactly where the annex says they go — cofacture doesn't check whether the code itself is valid.
- **No graphic representation (RIDE/PDF) generator.** DIAN doesn't validate this over SOAP; it's outside what this library does.
- **No orchestration layer.** Numbering, idempotency, retrying async operations, and persisting consecutive numbers are the caller's responsibility.

## Package map

| Package | Responsibility |
|---|---|
| [`domain`](./domain) | Plain Go structs for every document (`Invoice`, `CreditNote`, `DebitNote`, `AdjustmentNote`, `AttachedDocument`, `Event`, `Reclamo`, `Party`, `Tax`, `Line`, ...). No validation, no persistence. |
| [`builder`](./builder) | Assembles the UBL 2.1 + DIAN-extension XML tree from a domain model (`etree.Document`) — every document type plus the RADIAN `ApplicationResponse` events. Does not sign, hash, or send anything. |
| [`cufe`](./cufe) | Computes the CUFE (Invoice) per Technical Annex 1.9. |
| [`cude`](./cude) | Computes the CUDE for Credit Notes, Debit Notes, and Documento Equivalente Electrónico — same formula for all three, see the package's doc comment. |
| [`cuds`](./cuds) | Computes the CUDS (Support Document, Adjustment Note to the Support Document). |
| [`event`](./event) | Computes the CUDE for RADIAN events and holds their `ResponseCode` catalog and the Aceptación Tácita note template. |
| [`payroll`](./payroll) | Builds `NominaIndividual` XML and computes the CUNE. |
| [`securitycode`](./securitycode) | Computes `sts:SoftwareSecurityCode`. |
| [`qr`](./qr) | Builds the QR URL/content required in every document type's graphic representation. |
| [`signer`](./signer) | XAdES-EPES signing (C14N 1.0, RSA-SHA256) plus certificate/key loading (PEM and PKCS#12). |
| [`zip`](./zip) | Packages signed XML into the ZIP format and file-naming convention DIAN's receiving service requires. |
| [`soap`](./soap) | SOAP 1.2 + WS-Security client for `WcfDianCustomerServices` (habilitación and producción). |
| [`dian`](./dian) | Interprets DIAN's validation responses into a structured `Result`. |
| [`xml`](./xml) | Shared UBL/DIAN namespace constants. |

---

## Installation

```bash
go get github.com/diegofxm/cofacture
```

Requires Go 1.26.4 or newer (see `go.mod`).

---

## Quick start

The full pipeline for a single invoice, end to end. Error handling is abbreviated for readability — check every error in real code.

```go
// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package main

import (
	"log"
	"os"
	"time"

	"github.com/diegofxm/cofacture/builder"
	"github.com/diegofxm/cofacture/cufe"
	"github.com/diegofxm/cofacture/dian"
	"github.com/diegofxm/cofacture/domain"
	"github.com/diegofxm/cofacture/qr"
	"github.com/diegofxm/cofacture/securitycode"
	"github.com/diegofxm/cofacture/signer"
	"github.com/diegofxm/cofacture/soap"
	cfzip "github.com/diegofxm/cofacture/zip"
)

func main() {
	// 1. Load your DIAN-issued certificate (.p12) once, reuse it across documents.
	certBytes, _ := os.ReadFile("software.p12")
	cert, key, err := signer.LoadPKCS12(certBytes, "your-p12-password")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Build the domain model. Everything here comes from your own system
	//    (ERP, database, order service, etc.) — cofacture never reaches out to fetch it.
	//    supplierParty, customerParty, headerTaxes, totals, lines, numberingRange and
	//    softwareProvider below are placeholders for values you supply; numberingRange in
	//    particular comes from a prior call to soap.Client.GetNumberingRange.
	inv := domain.Invoice{
		ProfileID:         "DIAN 2.1",
		EnvironmentCode:   "2", // "1" production, "2" certification (habilitación)
		OperationTypeCode: "10",
		DocumentTypeCode:  "01",
		HashType:          "CUFE-SHA384",
		Prefix:            "SETP",
		Number:            "990000001",
		IssueDate:         "2026-08-27",
		IssueTime:         "10:15:00-05:00",
		CurrencyCode:      "COP",
		Supplier:          supplierParty,
		Customer:          customerParty,
		HeaderTaxes:       headerTaxes,
		Totals:            totals,
		Lines:             lines,
		NumberingRange:    numberingRange,
		SoftwareProvider:  softwareProvider,
	}

	// 3. Compute the identifiers DIAN requires before the document can be signed.
	//    technicalKey comes from GetNumberingRange; softwareID/pin are the credentials
	//    DIAN assigned when you activated your software.
	inv.CUFE = cufe.Compute(inv, technicalKey)
	inv.SoftwareSecurityCode = securitycode.Compute(softwareID, pin, inv.Prefix+inv.Number)
	inv.QRURL = qr.URL(inv.EnvironmentCode, inv.CUFE)

	// 4. Build the UBL XML tree.
	doc, err := builder.BuildInvoice(inv)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Sign it (XAdES-EPES).
	placeholder, err := builder.SignaturePlaceholder(doc)
	if err != nil {
		log.Fatal(err)
	}
	if err := signer.New(cert, key).Sign(doc.Root(), placeholder, "supplier", time.Now().In(domain.Bogota)); err != nil {
		log.Fatal(err)
	}
	xmlBytes, err := doc.WriteToBytes()
	if err != nil {
		log.Fatal(err)
	}

	// 6. Name and package the file the way DIAN's receiving service expects.
	fileName := cfzip.DocumentFileName(cfzip.KindInvoice, "900123456", cfzip.SoftwarePropioCode, 2026, 1)
	zipBytes, err := cfzip.Build([]cfzip.File{{Name: fileName, Content: xmlBytes}})
	if err != nil {
		log.Fatal(err)
	}

	// 7. Send it to DIAN and interpret the response.
	client := soap.New(soap.HabilitacionURL, cert, key)
	resp, err := client.SendBillSync(fileName, zipBytes)
	if err != nil {
		log.Fatal(err)
	}
	result, err := dian.Interpret(*resp)
	if err != nil {
		log.Fatal(err)
	}
	if !result.IsValid {
		log.Fatalf("DIAN rejected the invoice: %+v", result.Messages)
	}
	log.Println("Accepted:", result.StatusDescription)
}
```

Building a Support Document, Credit Note, Debit Note, Adjustment Note, or Payroll document follows the same shape — swap `builder.BuildInvoice` for `builder.BuildSupportDocument` / `builder.BuildCreditNote` / `builder.BuildDebitNote` / `builder.BuildAdjustmentNote` / `payroll.Build`, and `cufe.Compute` for `cuds.Compute` / `cude.Compute` / `payroll.Cune` as appropriate.

RADIAN events (`builder.BuildAcuseRecibo` and its four siblings) follow a related but not identical shape: `event.Compute` instead of `cufe.Compute`, `Sender`/`Receiver` instead of `Supplier`/`Customer`, and a `DocumentReference` pointing at the document the event applies to instead of a `BillingReference`. The event's QR is built from the *referenced* document's CUFE (`qr.URL(ev.EnvironmentCode, ev.DocumentReference.CUFE)`), not from the event's own CUDE.

A Documento Equivalente Electrónico (e.g. a POS ticket) is `builder.BuildInvoice` and `cude.Compute` again — same `domain.Invoice`, same pipeline as above — with five fields set differently: `ProfileID: "DIAN 2.1: Documento Equivalente POS"`, `OperationTypeCode: "10"` (this is what ends up as the `cbc:CustomizationID` XML element — the struct field is named after the DIAN concept it represents, not the element it renders to), `DocumentTypeCode: "20"` (the Technical Annex Documento Equivalente Electrónico V1.0, section 16.3, lists the codes for the other 9 sub-types), `HashType: "CUDE-SHA384"` (instead of `"CUFE-SHA384"`), and `softwarePIN` passed to `cude.Compute` where the invoice above passes `technicalKey` to `cufe.Compute`. Its Adjustment Notes are `builder.BuildCreditNote`/`BuildDebitNote` with `DocumentTypeCode`/`CreditNoteTypeCode` `"94"`/`"93"`, `OperationTypeCode` set to the *referenced* document's own type code (e.g. `"20"` for one adjusting a POS ticket), and `BillingReference.HashType` set to that referenced document's own hash scheme (`"CUDE-SHA384"` for a POS ticket, not `"CUFE-SHA384"`).

---

## Security notes

- **Never commit certificates, `.p12`/`.pfx` files, PINs, or software IDs.** `.gitignore` already excludes common patterns (`*.p12`, `*.pfx`, `*_cert.pem`, `*_key.pem`, `.env*`), but review `git status` before every commit regardless.
- `signer.LoadPEM` / `signer.LoadPKCS12` load key material into memory only — the library never persists it anywhere.
- `securitycode.Compute` takes your DIAN-assigned `softwareID`/`pin` directly; treat both as secrets with the same care as a private key.
- `go test ./...` inside this module never touches the network or requires credentials.

---

## Testing

```bash
go test ./...
```

Runs the unit test suite — XML golden-file comparisons, hash/signature vector checks, response-parsing tests — with no network access and no credentials required.

Tests that submit documents to DIAN's certification (habilitación) environment live in a separate sibling project outside this repository, since they require a real certificate and DIAN credentials that must never be checked into source control.

---

## Contributing

Issues and pull requests are welcome. Found a bug? Please open an issue.

---

## Disclaimer

This project is an independent, community-built toolkit. It is **not affiliated with, endorsed by, or officially certified by DIAN** (*Dirección de Impuestos y Aduanas Nacionales*). Achieving DIAN's *habilitación* (certification) for a specific NIT/software combination is a separate process the taxpayer/technology provider must complete directly with DIAN; this library helps you build the documents involved, but using it correctly does not by itself grant certification.

---

## License

Cofacture is **dual-licensed**.

### Open Source — AGPL-3.0

Cofacture is available under the terms of the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

You are free to use, study, modify, and distribute cofacture in accordance with the terms and conditions of the AGPL-3.0.

The full license text is available in the [`AGPL-3.0`](LICENSE) file.

### Commercial License

If you want to use Cofacture in a **proprietary, closed-source, or commercial product** without being subject to the obligations of the AGPL-3.0, a separate **commercial license** is available.

A commercial license may be appropriate for companies or organizations that:

- Integrate Cofacture into proprietary or closed-source software.
- Distribute products containing Cofacture under proprietary terms.
- Do not want to release their modifications or combined work under the AGPL-3.0.
- Require commercial licensing terms, warranties, support, or other contractual arrangements.

For information about commercial licensing, please contact:

**Diego Montoya**  
GitHub: [@diegofxm](https://github.com/diegofxm)

Unless you have obtained a separate commercial license, use of Cofacture is governed by the **AGPL-3.0**.

### Copyright

Copyright © 2026 Diego Montoya.

Go to [`AGPL-3.0`](https://www.gnu.org/licenses/agpl-3.0.html) for the complete AGPL-3.0 license terms.