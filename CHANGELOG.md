# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/).

## [0.2.0] — 2026-09-04

### Changed

- **Breaking:** renamed `domain.Invoice` to `domain.Document` (file `domain/invoice.go` →
  `domain/document.go`). It was never Invoice-only — the same struct is the model for Support
  Document and Documento Equivalente Electrónico, and is embedded in `CreditNote`, `DebitNote`,
  and `AdjustmentNote`. The new name and doc comment reflect that; callers using `domain.Invoice`
  or the `.Invoice` embedded-field selector need to update to `domain.Document`/`.Document`.

## [0.1.2] — 2026-09-04

### Fixed

- Corrected a stale function-name reference in a doc comment (`appendSupplierParty` →
  `appendDSSupplierParty`).

### Testing

- Added regression tests for `signer.LoadPKCS12`'s CA-chain fix, and for `LoadPEM` — both were
  previously untested.
- Added tests for `qr.SupportDocumentURL`, `qr.AdjustmentNoteContent`, and the tax-fallback
  branches of both QR content builders (coverage 48.8% → 95.1%).
- Added tests for `internal/dianhash.Seed`, previously untested (coverage 0% → 100%).

## [0.1.1] — 2026-09-04

### Fixed

- Corrected the expected `ProfileID` value for Invoice, Credit Note, and Debit Note — found
  while testing against DIAN's certification environment. The generic `"DIAN 2.1"` does not
  match what the Technical Annex expects in `cbc:ProfileID`; it must name the specific document,
  e.g. `"DIAN 2.1: Factura Electrónica de Venta"`, `"DIAN 2.1: Nota Crédito de Factura
  Electrónica de Venta"`, or `"DIAN 2.1: Nota Débito de Factura Electrónica de Venta"`.
  `ProfileID` is a caller-supplied field on `domain.Invoice`, not hardcoded by the builder, so
  callers who used the old `"DIAN 2.1"` value should update it.
- `signer.LoadPKCS12` now accepts `.p12`/`.pfx` files that bundle a CA chain alongside the leaf
  certificate — the previous implementation (`pkcs12.Decode`) rejected any file with more than
  the certificate and key ("expected exactly two safe bags"), which real DIAN-issued certificates
  routinely violate. Switched to `pkcs12.DecodeChain`; the chain itself is discarded, since DIAN's
  XAdES-EPES signature only ever embeds the leaf certificate.

### Documentation

- Documented a DIAN business rule confirmed against real submissions: the issuer's identification
  must use `TypeCode: "31"` (NIT) — not `"13"` (cédula) — for Debit Notes, and the acquirer's must
  do the same for Support Documents, even when the same party identifies with `"13"` without
  issue on an Invoice or Credit Note. Also documented that `TaxSchemeCode`/`TaxSchemeName` must
  reflect a real tax regime once a party is NIT-identified. Doc comments only, no behavior change.
- Added a "Document coverage" table to the README showing certification status per document
  type, and updated the Quick Start example to fetch the numbering range via
  `GetNumberingRange`. General wording cleanup throughout.

## [0.1.0] — 2026-09-04

Initial release.

### Added

- UBL 2.1 document builders: Electronic Sales Invoice, Credit Note, Debit Note, Support
  Document, Adjustment Note to the Support Document, Attached Document, and Documento
  Equivalente Electrónico (POS ticket and other sub-types).
- Individual Electronic Payroll builder (`NominaIndividual`) and its adjustment/cancellation
  variants.
- RADIAN event builders: Acuse de Recibo, Reclamo, Recibo del Bien, Aceptación Expresa,
  Aceptación Tácita.
- CUFE, CUDE, CUDS, and CUNE computation per the DIAN Technical Annexes.
- XAdES-EPES signing (C14N 1.0, RSA-SHA256), loading certificates from PEM or PKCS#12.
- ZIP packaging matching the file-naming convention DIAN's receiving service requires.
- SOAP 1.2 + WS-Security client implementing all 16 operations of the `WcfDianCustomerServices`
  contract, for both habilitación and producción.
- Response parser turning DIAN's validation output into a structured result.

[0.2.0]: https://github.com/diegofxm/cofacture/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/diegofxm/cofacture/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/diegofxm/cofacture/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/diegofxm/cofacture/releases/tag/v0.1.0
