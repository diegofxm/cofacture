// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package payroll builds, signs and sends Individual Electronic Payroll documents
// (NominaIndividual) to DIAN's web service (SendNominaSync).
//
// It does not create its own SOAP clients — it delegates to cofacture/soap, whose Client
// already knows how to talk to WcfDianCustomerServices. It does contain its own XML builder
// because NominaIndividual's structure is not UBL 2.1 and does not share the builder/ helpers.
//
// Field names below intentionally mirror DIAN's own Spanish XML attribute names
// (NominaIndividualElectronicaXSD.xsd) so the mapping between this struct and the wire format
// stays obvious — only the doc comments are in English.
package payroll

// Nomina gathers all the data needed to generate, sign and send a NominaIndividual or
// NominaIndividualDeAjuste.
type Nomina struct {
	// Sequence and type.
	Consecutivo string // e.g. "1"
	Prefijo     string // e.g. "NE"
	Numero      string // full number = Prefijo + Consecutivo

	TipoXML string // "102" NominaIndividual, "103" Adjustment-Replace, "104" Adjustment-Delete

	// CUNENovedad is the CUNE of the original NominaIndividual this document adjusts. Required
	// when TipoXML is "103" (Adjustment-Replace) or "104" (Adjustment-Delete); Build rejects an
	// adjustment with this left empty. Ignored (and must be left empty) for a normal "102"
	// payroll — DIAN's Novedad/@CUNENov pair only applies to adjustments.
	CUNENovedad string

	// DIAN environment: "1" production, "2" certification/testing.
	Ambiente string

	// Period and document-generation dates/times.
	FechaGen               string // "2024-01-31"
	HoraGen                string // "09:00:00-05:00"
	FechaIngreso           string // "2020-01-01"
	FechaRetiro            string // "" empty = not applicable
	FechaLiquidacionInicio string // "2024-01-01"
	FechaLiquidacionFin    string // "2024-01-31"
	TiempoLaborado         int    // days worked in the period

	// Place of generation.
	Pais               string // "CO"
	DepartamentoEstado string // DIAN code, e.g. "11"
	MunicipioCiudad    string // DIAN code, e.g. "001"
	Idioma             string // ISO 639-1, e.g. "es"

	// Document metadata.
	PeriodoNomina int    // 1=Weekly 2=Ten-day 3=Biweekly 4=Monthly ...
	TipoMoneda    string // "COP"
	TRM           string // "1.00" (exchange rate; "1.00" for COP→COP)
	Notas         string // optional free text

	// Own software / Technology Provider (PT).
	SoftwareID string // UUID registered with DIAN
	PIN        string // numeric software PIN

	// XML provider (who generates the document — company or PT).
	// For natural persons use PrimerApellido/PrimerNombre; for companies use RazonSocial.
	ProveedorNIT         string
	ProveedorDV          string
	ProveedorRazonSocial string
	ProveedorApellido1   string
	ProveedorApellido2   string
	ProveedorNombre1     string
	ProveedorNombre2     string // OtrosNombres

	// Payroll payment dates.
	FechasPago []string // []"2024-01-31"

	Empleador   Empleador
	Trabajador  Trabajador
	Pago        Pago
	Devengados  Devengados
	Deducciones Deducciones

	// Totals (if zero, they are computed from Devengados/Deducciones in Build).
	RedondeoCents         int64
	DevengadosTotalCents  int64
	DeduccionesTotalCents int64
	ComprobanteTotalCents int64
}

// Empleador identifies the company or person paying the payroll.
// For natural persons use PrimerApellido/PrimerNombre/etc.; for companies use RazonSocial.
type Empleador struct {
	RazonSocial     string
	PrimerApellido  string
	SegundoApellido string
	PrimerNombre    string
	OtrosNombres    string
	NIT             string
	DV              string
	Pais            string // "CO"
	Departamento    string // DIAN code
	Municipio       string // DIAN code
	Direccion       string
}

// Trabajador identifies the employee.
type Trabajador struct {
	TipoTrabajador    string // "01" employee, etc.
	SubTipoTrabajador string // "00" not applicable, "01" dependent, etc.
	AltoRiesgoPension bool
	TipoDocumento     string // "13" national ID, "31" NIT, etc.
	NumeroDocumento   string
	PrimerApellido    string
	SegundoApellido   string
	PrimerNombre      string
	OtrosNombres      string
	LugarPais         string // "CO"
	LugarDepartamento string // DIAN code
	LugarMunicipio    string // DIAN code
	LugarDireccion    string
	SalarioIntegral   bool
	TipoContrato      string // "1" fixed-term, "2" indefinite, "3" apprenticeship, "4" internship
	SueldoCents       int64  // base monthly salary
	CodigoTrabajador  string // internal employee code
}

// Pago describes the payroll's payment method to the worker.
type Pago struct {
	Forma        string // "1" cash, "2" credit
	Metodo       string // "10" cash, "42" bank transfer, etc.
	Banco        string
	TipoCuenta   string // "AHORRO" (savings), "CORRIENTE" (checking)
	NumeroCuenta string
}

// Devengados groups the worker's earnings for the period.
type Devengados struct {
	Basico     Basico
	Transporte *Transporte // nil = not applicable
}

// Basico is the base salary for the period (required).
type Basico struct {
	DiasTrabajados       int
	SueldoTrabajadoCents int64
}

// Transporte covers the transportation allowance and per diems.
type Transporte struct {
	AuxilioTransporteCents int64
	ViaticoManuAlojSCents  int64 // salary-affecting
	ViaticoManuAlojNSCents int64 // non-salary-affecting
}

// Deducciones groups the worker's deductions for the period.
type Deducciones struct {
	Salud                *DeduccionPct // nil = not applicable
	FondoPension         *DeduccionPct // nil = not applicable
	FondoSP              *FondoSP      // nil = not applicable (only for salaries > 4 SMLMV)
	RetencionFuenteCents int64
}

// DeduccionPct is a deduction defined by percentage and amount.
type DeduccionPct struct {
	Porcentaje     float64
	DeduccionCents int64
}

// FondoSP is the deduction for the Solidarity Pension Fund (Fondo de Solidaridad Pensional).
type FondoSP struct {
	Porcentaje        float64
	DeduccionSPCents  int64
	PorcentajeSub     float64
	DeduccionSubCents int64
}
