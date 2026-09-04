# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/).

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

[0.1.0]: https://github.com/diegofxm/cofacture/releases/tag/v0.1.0
