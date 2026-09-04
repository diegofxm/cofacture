// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package payroll

import "testing"

// validNomina returns a minimal but complete Nomina for a normal ("102") payroll — enough to
// get through Build() without hitting an unrelated validation error, so each test below only
// needs to override the fields it actually cares about.
func validNomina() Nomina {
	return Nomina{
		Consecutivo:            "1",
		Prefijo:                "NE",
		Numero:                 "NE1",
		TipoXML:                "102",
		Ambiente:               "2",
		FechaGen:               "2024-01-31",
		HoraGen:                "09:00:00-05:00",
		FechaIngreso:           "2020-01-01",
		FechaLiquidacionInicio: "2024-01-01",
		FechaLiquidacionFin:    "2024-01-31",
		TiempoLaborado:         31,
		Pais:                   "CO",
		DepartamentoEstado:     "11",
		MunicipioCiudad:        "001",
		Idioma:                 "es",
		PeriodoNomina:          4,
		TipoMoneda:             "COP",
		TRM:                    "1.00",
		SoftwareID:             "12345678-1234-1234-1234-123456789012",
		ProveedorNIT:           "700085371",
		ProveedorDV:            "1",
		ProveedorRazonSocial:   "Proveedor SAS",
		FechasPago:             []string{"2024-01-31"},
		Empleador: Empleador{
			RazonSocial:  "Empleador SAS",
			NIT:          "800199436",
			DV:           "1",
			Pais:         "CO",
			Departamento: "11",
			Municipio:    "001",
			Direccion:    "Calle 1 # 2-3",
		},
		Trabajador: Trabajador{
			TipoTrabajador:    "01",
			SubTipoTrabajador: "00",
			TipoDocumento:     "13",
			NumeroDocumento:   "123456789",
			PrimerApellido:    "Perez",
			PrimerNombre:      "Juan",
			LugarPais:         "CO",
			LugarDepartamento: "11",
			LugarMunicipio:    "001",
			LugarDireccion:    "Calle 1 # 2-3",
			TipoContrato:      "1",
			SueldoCents:       350_000_000,
			CodigoTrabajador:  "1",
		},
		Pago: Pago{
			Forma:        "1",
			Metodo:       "42",
			Banco:        "Banco de Prueba",
			TipoCuenta:   "AHORRO",
			NumeroCuenta: "123456789",
		},
		Devengados: Devengados{
			Basico: Basico{DiasTrabajados: 31, SueldoTrabajadoCents: 350_000_000},
		},
		DevengadosTotalCents:  350_000_000,
		DeduccionesTotalCents: 0,
		ComprobanteTotalCents: 350_000_000,
	}
}

// TestBuild_NormalPayroll_NovedadIsFalse confirms a regular "102" NominaIndividual serializes
// Novedad="false" with an empty CUNENov, and does not require CUNENovedad to be set.
func TestBuild_NormalPayroll_NovedadIsFalse(t *testing.T) {
	n := validNomina() // TipoXML "102", CUNENovedad left empty

	doc, err := Build(n, "cune-placeholder", "sc-placeholder", "https://example.test/qr")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	novedad := doc.FindElement("//Novedad")
	if novedad == nil {
		t.Fatal("Novedad element not found in the built document")
	}
	if got := novedad.Text(); got != "false" {
		t.Errorf("Novedad text = %q, want %q", got, "false")
	}
	if got := novedad.SelectAttrValue("CUNENov", "MISSING"); got != "" {
		t.Errorf("CUNENov = %q, want empty", got)
	}
}

// TestBuild_Adjustment_RequiresCUNENovedad confirms Build rejects an adjustment payroll
// ("103"/"104") that doesn't carry the original document's CUNE — this is exactly the gap that
// previously made adjustment payroll unusable (Novedad was hardcoded to "false").
func TestBuild_Adjustment_RequiresCUNENovedad(t *testing.T) {
	for _, tipoXML := range []string{"103", "104"} {
		n := validNomina()
		n.TipoXML = tipoXML
		n.CUNENovedad = ""

		if _, err := Build(n, "cune-placeholder", "sc-placeholder", "https://example.test/qr"); err == nil {
			t.Errorf("TipoXML %q: Build should fail when CUNENovedad is empty", tipoXML)
		}
	}
}

// TestBuild_Adjustment_SetsNovedadTrue confirms an adjustment payroll with CUNENovedad set
// serializes Novedad="true" and carries the original CUNE in the CUNENov attribute.
func TestBuild_Adjustment_SetsNovedadTrue(t *testing.T) {
	const originalCUNE = "16560dc8956122e84ffb743c817fe7d494e058a44d9ca3fa4c234c268b4f766003253fbee7ea4af9682dd57210f3bac2"

	n := validNomina()
	n.TipoXML = "103"
	n.CUNENovedad = originalCUNE

	doc, err := Build(n, "cune-placeholder", "sc-placeholder", "https://example.test/qr")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	novedad := doc.FindElement("//Novedad")
	if novedad == nil {
		t.Fatal("Novedad element not found in the built document")
	}
	if got := novedad.Text(); got != "true" {
		t.Errorf("Novedad text = %q, want %q", got, "true")
	}
	if got := novedad.SelectAttrValue("CUNENov", "MISSING"); got != originalCUNE {
		t.Errorf("CUNENov = %q, want %q", got, originalCUNE)
	}
}
