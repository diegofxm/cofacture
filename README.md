# cofacture

[![Go Reference](https://pkg.go.dev/badge/github.com/diegofxm/cofacture.svg)](https://pkg.go.dev/github.com/diegofxm/cofacture)
[![License](https://img.shields.io/badge/license-AGPL--3.0-blue)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/diegofxm/cofacture)](go.mod)

A Go library for Colombian DIAN electronic invoicing: UBL 2.1 document generation, XAdES-EPES digital signing, CUFE/CUDE/CUDS/CUNE computation, and a SOAP client for DIAN's `WcfDianCustomerServices` web services.

`cofacture` is a **library, not a platform** — no database, no HTTP server, no opinion about how you store your data. You feed it plain Go structs; it hands back signed XML, ready to package and send. You own the pipeline (build → hash → sign → zip → send): nothing here retries on your behalf, persists a consecutive number, or validates a catalog code before putting it where the technical annex says it goes.

---

## Features

- **Full document coverage** — Electronic Sales Invoice, Credit Note, Debit Note, Support Document, Adjustment Note to the Support Document, Attached Document, the five RADIAN acceptance/rejection events, Individual Electronic Payroll, and Documento Equivalente Electrónico (POS ticket and its other sub-types).
- **XAdES-EPES signing** — inclusive C14N 1.0 canonicalization, RSA-SHA256, DIAN's fixed signature policy, built from PEM or PKCS#12 (`.p12`/`.pfx`) certificates.
- **Hash formulas verified against DIAN's own published worked examples** (Technical Annex 1.9), not just internal regression tests — CUFE, CUDE, CUDS, and CUNE each have a dedicated test reproducing DIAN's official example.
- **SOAP 1.2 + WS-Security client** implementing 16 operations of the `WcfDianCustomerServices` contract: `SendBillSync`, `SendBillAsync`, `SendBillAttachmentAsync`, `SendTestSetAsync`, `GetStatus`, `GetStatusZip`, `GetNumberingRange`, `SendNominaSync`, `SendNominaSyncTestSet`, `SendEventUpdateStatus`, `GetStatusEvent`, `GetAcquirer`, `GetXmlByDocumentKey`, `GetReferenceNotes`, `GetDocumentInfo`, `GetExchangeEmails`.
- **Response interpretation** — DIAN's validation messages are parsed into a structured `Result` (rejections vs. informational notices), ready to branch on.

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
	"strconv"
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
	// 1. Load your DIAN-issued certificate (.p12) once, reuse it across documents — for both
	//    XAdES document signing and the SOAP WS-Security header.
	certBytes, _ := os.ReadFile("software.p12")
	cert, key, err := signer.LoadPKCS12(certBytes, "your-p12-password")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Query the numbering range DIAN authorized for your issuer + software pair (cache the
	//    result — it doesn't change per document). technicalKey feeds cufe.Compute() below.
	client := soap.New(soap.HabilitacionURL, cert, key)
	ranges, err := client.GetNumberingRange("900123456", "900123456", softwareUUID)
	if err != nil {
		log.Fatal(err)
	}
	nr := ranges.ResponseList[0]

	// 3. Build the domain model. Everything here comes from your own system
	//    (ERP, database, order service, etc.) — cofacture never reaches out to fetch it.
	//    supplierParty, customerParty, headerTaxes, totals, lines and softwareProvider below
	//    are placeholders for values you supply.
	inv := domain.Invoice{
		ProfileID:         "DIAN 2.1: Factura Electrónica de Venta",
		EnvironmentCode:   "2", // "1" production, "2" certification (habilitación)
		OperationTypeCode: "10",
		DocumentTypeCode:  "01",
		HashType:          "CUFE-SHA384",
		Prefix:            nr.Prefix,
		Number:            "990000001",
		IssueDate:         time.Now().Format("2006-01-02"),
		IssueTime:         time.Now().Format("15:04:05-07:00"),
		CurrencyCode:      "COP",
		Supplier:          supplierParty,
		Customer:          customerParty,
		HeaderTaxes:       headerTaxes,
		Totals:            totals,
		Lines:             lines,
		NumberingRange: domain.NumberingRange{
			AuthorizedCode: nr.ResolutionNumber,
			Prefix:         nr.Prefix,
			StartNumber:    strconv.FormatInt(nr.FromNumber, 10),
			EndNumber:      strconv.FormatInt(nr.ToNumber, 10),
			StartDate:      nr.ValidDateFrom,
			EndDate:        nr.ValidDateTo,
		},
		SoftwareProvider: softwareProvider,
	}

	// 4. Compute the identifiers DIAN requires before the document can be signed.
	//    softwareID/pin are the credentials DIAN assigned when you activated your software.
	inv.CUFE = cufe.Compute(inv, nr.TechnicalKey)
	inv.SoftwareSecurityCode = securitycode.Compute(softwareID, pin, inv.Prefix+inv.Number)
	inv.QRURL = qr.URL(inv.EnvironmentCode, inv.CUFE)

	// 5. Build the UBL XML tree and sign it (XAdES-EPES).
	doc, err := builder.BuildInvoice(inv)
	if err != nil {
		log.Fatal(err)
	}
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
	fileName := cfzip.DocumentFileName(cfzip.KindInvoice, "900123456", cfzip.SoftwarePropioCode, time.Now().Year(), 1)
	zipBytes, err := cfzip.Build([]cfzip.File{{Name: fileName, Content: xmlBytes}})
	if err != nil {
		log.Fatal(err)
	}

	// 7. Send it to DIAN and interpret the response.
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

Every other document type follows the same shape — build the domain model, compute its hash (`cude.Compute`, `cuds.Compute`, `event.Compute`, or `payroll.Cune` in place of `cufe.Compute`), build the XML tree with the matching builder (`builder.BuildSupportDocument` / `builder.BuildCreditNote` / `builder.BuildDebitNote` / `builder.BuildAdjustmentNote` / `payroll.Build`), sign, package, and send. See [Document coverage](#document-coverage) below for the full list and [Package map](#package-map) for where each piece lives.

RADIAN events (`builder.BuildAcuseRecibo` and its four siblings) follow a related but not identical shape: `event.Compute` instead of `cufe.Compute`, `Sender`/`Receiver` instead of `Supplier`/`Customer`, and a `DocumentReference` pointing at the document the event applies to instead of a `BillingReference`. The event's QR is built from the *referenced* document's CUFE (`qr.URL(ev.EnvironmentCode, ev.DocumentReference.CUFE)`), not from the event's own CUDE.

A Documento Equivalente Electrónico (e.g. a POS ticket) is `builder.BuildInvoice` and `cude.Compute` again — same `domain.Invoice`, same pipeline as above — with five fields set differently: `ProfileID: "DIAN 2.1: Documento Equivalente POS"`, `OperationTypeCode: "10"` (this is what ends up as the `cbc:CustomizationID` XML element — the struct field is named after the DIAN concept it represents, not the element it renders to), `DocumentTypeCode: "20"` (the Technical Annex Documento Equivalente Electrónico V1.0, section 16.3, lists the codes for the other 9 sub-types), `HashType: "CUDE-SHA384"` (instead of `"CUFE-SHA384"`), and `softwarePIN` passed to `cude.Compute` where the invoice above passes `technicalKey` to `cufe.Compute`. Its Adjustment Notes are `builder.BuildCreditNote`/`BuildDebitNote` with `DocumentTypeCode`/`CreditNoteTypeCode` `"94"`/`"93"`, `OperationTypeCode` set to the *referenced* document's own type code (e.g. `"20"` for one adjusting a POS ticket), and `BillingReference.HashType` set to that referenced document's own hash scheme (`"CUDE-SHA384"` for a POS ticket, not `"CUFE-SHA384"`).

---

## Document coverage

| Document | Type code(s) | DIAN certification status |
|---|---|---|
| Electronic Sales Invoice | `01` | Confirmed accepted in DIAN's certification (habilitación) environment |
| Credit Note | `91` | Hash formula matches DIAN's official worked example |
| Debit Note | `92` | Hash formula matches DIAN's official worked example |
| Support Document | `05` | Built, signed, and submitted to DIAN's certification environment during testing |
| Adjustment Note to the Support Document | `95` | Built, signed, and submitted to DIAN's certification environment during testing |
| Attached Document (container for Invoice/Credit Note/Debit Note) | — | Built and independently signature-verified; delivered to the acquirer, not submitted to DIAN by design |
| RADIAN events — Acuse de Recibo, Reclamo, Recibo del Bien, Aceptación Expresa, Aceptación Tácita | `030`–`034` | Built per the technical annex's field tables; not yet submitted |
| Individual Electronic Payroll & Adjustment | `102`/`103`/`104` | Built and independently signature-verified; not yet submitted |
| Documento Equivalente Electrónico (POS ticket, etc.) | `20`, `25`, ... | Supported via the Invoice/Credit Note/Debit Note builders with the applicable type code; not yet submitted |

### Design boundaries (not gaps)

- **No catalog/reference-data validation** — tax types, unit codes, city/DANE codes, payment methods, etc. `cofacture` trusts the caller and only knows where a code goes in the XML, not whether it's valid.
- **No graphic representation (RIDE/PDF) generator.** DIAN doesn't validate this over SOAP; it's outside this library's scope.
- **No orchestration layer.** No single "send a document" call — you own numbering, idempotency, retry logic, and persisting consecutive numbers.

---

## Package map

| Package | Responsibility |
|---|---|
| [`domain`](./domain) | Plain Go structs for every document (`Invoice`, `CreditNote`, `DebitNote`, `AdjustmentNote`, `AttachedDocument`, `Event`, `Reclamo`, `Party`, `Tax`, `Line`, ...). No validation, no persistence. |
| [`builder`](./builder) | Assembles the UBL 2.1 + DIAN-extension XML tree from a domain model (`etree.Document`) — every document type plus the RADIAN `ApplicationResponse` events. Does not sign, hash, or send anything. |
| [`cufe`](./cufe) | Computes the CUFE for the Electronic Sales Invoice. |
| [`cude`](./cude) | Computes the CUDE for Credit Notes, Debit Notes, and Documento Equivalente Electrónico — same formula for all three, see the package's doc comment. |
| [`cuds`](./cuds) | Computes the CUDS for the Support Document and its Adjustment Note. |
| [`event`](./event) | Computes the CUDE for RADIAN events and holds their `ResponseCode` catalog and the Aceptación Tácita note template. |
| [`payroll`](./payroll) | Builds `NominaIndividual` XML (a distinct, non-UBL schema) and computes the CUNE. |
| [`securitycode`](./securitycode) | Computes `sts:SoftwareSecurityCode`. |
| [`qr`](./qr) | Builds the QR URL/content required in each document type's graphic representation. |
| [`signer`](./signer) | XAdES-EPES signing (C14N 1.0, RSA-SHA256) plus certificate/key loading (PEM and PKCS#12). |
| [`zip`](./zip) | Packages signed XML into the ZIP format and file-naming convention DIAN's receiving service requires. |
| [`soap`](./soap) | SOAP 1.2 + WS-Security client for `WcfDianCustomerServices` (habilitación and producción). |
| [`dian`](./dian) | Interprets DIAN's validation responses into a structured `Result` (rejections vs. notices, embedded `ApplicationResponse`, etc.). |
| [`xml`](./xml) | Shared UBL/DIAN namespace constants. |

---

## Security notes

- **Never commit certificates, `.p12`/`.pfx` files, PINs, or software IDs.** `.gitignore` already excludes common patterns (`*.p12`, `*.pfx`, `*_cert.pem`, `*_key.pem`, `.env*`), but review `git status` before every commit regardless.
- `signer.LoadPEM` / `signer.LoadPKCS12` load key material into memory only — this library never persists it anywhere.
- `securitycode.Compute` takes your DIAN-assigned `softwareID`/`pin` directly; treat both as secrets with the same care as a private key.
- Tests that talk to DIAN's real certification server require a real certificate and credentials that are never part of this module.

---

## Testing

```bash
go test ./...
```

Runs the full unit test suite — XML golden-file comparisons, hash/signature vector checks (several cross-checked against DIAN's own published worked examples), and response-parsing tests — with no network access and no credentials required.

Tests that submit documents to DIAN's certification (habilitación) environment live in a separate sibling project outside this module, since they require a real certificate and DIAN credentials that must never be checked into source control.

---

## Contributing

Issues and pull requests are welcome. Please include a clear description of the problem or feature, and add or update tests for any behavior change.

---

## Disclaimer

This project is an independent, community-built toolkit. It is **not affiliated with, endorsed by, or officially certified by DIAN** (*Dirección de Impuestos y Aduanas Nacionales*). Achieving DIAN's *habilitación* (certification) for a specific NIT/software combination is a separate process the taxpayer/technology provider must complete directly with DIAN; this library helps you build the documents involved, but using it correctly does not by itself grant certification. See [Document coverage](#document-coverage) for exactly which document types have been confirmed against a real DIAN server before relying on this library in production.

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