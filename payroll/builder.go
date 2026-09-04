// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package payroll

import (
	"fmt"
	"strconv"

	"github.com/beevik/etree"
	"github.com/diegofxm/cofacture/domain"
	ubl "github.com/diegofxm/cofacture/xml"
)

const (
	nsNomina             = "dian:gov:co:facturaelectronica:NominaIndividual"
	schemaLocationNomina = "dian:gov:co:facturaelectronica:NominaIndividual NominaIndividualElectronicaXSD.xsd"

	// EncripCUNE is the fixed value of the InformacionGeneral/@EncripCUNE attribute.
	EncripCUNE = "CUNE-SHA384"
	// Version is the fixed value of the InformacionGeneral/@Version attribute.
	Version = "V1.0: Documento Soporte de Pago de Nómina Electrónica"
)

// Build builds the XML tree of a NominaIndividual, with CUNE and CodigoQR already computed.
// The document is not signed yet — call builder.SignaturePlaceholder + signer.Sign afterward,
// same pipeline as invoices.
//
// The caller is responsible for computing:
//   - n.DevengadosTotal / n.DeduccionesTotal / n.ComprobanteTotal (if not already computed)
//   - The CUNE via Cune() and assigning it to n before calling Build
//   - The SoftwareSC via securitycode.Compute(softwareID, pin, numDoc) and assigning it to n
//   - The CodigoQR via qr.URL(ambiente, cune) and passing it as a parameter
func Build(n Nomina, cune, softwareSC, codigoQR string) (*etree.Document, error) {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8" standalone="no"`)

	root := doc.CreateElement("NominaIndividual")
	root.CreateAttr("xmlns", nsNomina)
	root.CreateAttr("xmlns:xs", ubl.NSXsi)
	root.CreateAttr("xmlns:ds", ubl.NSDs)
	root.CreateAttr("xmlns:ext", ubl.NSExt)
	root.CreateAttr("xmlns:xades", ubl.NSXades)
	root.CreateAttr("xmlns:xades141", ubl.NSXades141)
	root.CreateAttr("xmlns:xsi", ubl.NSXsi)
	root.CreateAttr("SchemaLocation", "")
	root.CreateAttr("xsi:schemaLocation", schemaLocationNomina)

	// ext:UBLExtensions — just an empty ext:UBLExtension reserved for the signature.
	// (NominaIndividual has no DianExtensions: its equivalents are ProveedorXML, CodigoQR, etc.)
	root.CreateElement("ext:UBLExtensions").
		CreateElement("ext:UBLExtension").
		CreateElement("ext:ExtensionContent")

	// Novedad marks this document as an adjustment ("novedad") of a previously submitted
	// NominaIndividual — true for TipoXML "103"/"104", false for a normal "102" payroll.
	// CUNENov must carry the original document's CUNE when this is an adjustment.
	isAdjustment := n.TipoXML == "103" || n.TipoXML == "104"
	if isAdjustment && n.CUNENovedad == "" {
		return nil, fmt.Errorf("payroll: CUNENovedad is required when TipoXML is an adjustment (\"103\"/\"104\")")
	}
	novedad := root.CreateElement("Novedad")
	novedad.CreateAttr("CUNENov", n.CUNENovedad)
	novedad.SetText(boolStr(isAdjustment))

	// Periodo
	fecRetiro := n.FechaRetiro
	if fecRetiro == "" {
		fecRetiro = "9999-12-31"
	}
	periodo := root.CreateElement("Periodo")
	periodo.CreateAttr("FechaIngreso", n.FechaIngreso)
	periodo.CreateAttr("FechaRetiro", fecRetiro)
	periodo.CreateAttr("FechaLiquidacionInicio", n.FechaLiquidacionInicio)
	periodo.CreateAttr("FechaLiquidacionFin", n.FechaLiquidacionFin)
	periodo.CreateAttr("TiempoLaborado", strconv.Itoa(n.TiempoLaborado))
	periodo.CreateAttr("FechaGen", n.FechaGen)

	// NumeroSecuenciaXML
	numSeq := root.CreateElement("NumeroSecuenciaXML")
	numSeq.CreateAttr("CodigoTrabajador", n.Trabajador.CodigoTrabajador)
	numSeq.CreateAttr("Prefijo", n.Prefijo)
	numSeq.CreateAttr("Consecutivo", n.Consecutivo)
	numSeq.CreateAttr("Numero", n.Numero)

	// LugarGeneracionXML
	lugar := root.CreateElement("LugarGeneracionXML")
	lugar.CreateAttr("Pais", n.Pais)
	lugar.CreateAttr("DepartamentoEstado", n.DepartamentoEstado)
	lugar.CreateAttr("MunicipioCiudad", n.MunicipioCiudad)
	lugar.CreateAttr("Idioma", n.Idioma)

	// ProveedorXML
	proveedor := root.CreateElement("ProveedorXML")
	proveedor.CreateAttr("RazonSocial", n.ProveedorRazonSocial)
	proveedor.CreateAttr("PrimerApellido", n.ProveedorApellido1)
	proveedor.CreateAttr("SegundoApellido", n.ProveedorApellido2)
	proveedor.CreateAttr("PrimerNombre", n.ProveedorNombre1)
	proveedor.CreateAttr("OtrosNombres", n.ProveedorNombre2)
	proveedor.CreateAttr("NIT", n.ProveedorNIT)
	proveedor.CreateAttr("DV", n.ProveedorDV)
	proveedor.CreateAttr("SoftwareID", n.SoftwareID)
	proveedor.CreateAttr("SoftwareSC", softwareSC)

	// CodigoQR
	root.CreateElement("CodigoQR").SetText(codigoQR)

	// InformacionGeneral
	info := root.CreateElement("InformacionGeneral")
	info.CreateAttr("Version", Version)
	info.CreateAttr("Ambiente", n.Ambiente)
	info.CreateAttr("TipoXML", n.TipoXML)
	info.CreateAttr("CUNE", cune)
	info.CreateAttr("EncripCUNE", EncripCUNE)
	info.CreateAttr("FechaGen", n.FechaGen)
	info.CreateAttr("HoraGen", n.HoraGen)
	info.CreateAttr("PeriodoNomina", strconv.Itoa(n.PeriodoNomina))
	info.CreateAttr("TipoMoneda", n.TipoMoneda)
	info.CreateAttr("TRM", n.TRM)

	// Notas (optional)
	if n.Notas != "" {
		root.CreateElement("Notas").SetText(n.Notas)
	}

	// Empleador
	emp := root.CreateElement("Empleador")
	emp.CreateAttr("RazonSocial", n.Empleador.RazonSocial)
	emp.CreateAttr("PrimerApellido", n.Empleador.PrimerApellido)
	emp.CreateAttr("SegundoApellido", n.Empleador.SegundoApellido)
	emp.CreateAttr("PrimerNombre", n.Empleador.PrimerNombre)
	emp.CreateAttr("OtrosNombres", n.Empleador.OtrosNombres)
	emp.CreateAttr("NIT", n.Empleador.NIT)
	emp.CreateAttr("DV", n.Empleador.DV)
	emp.CreateAttr("Pais", n.Empleador.Pais)
	emp.CreateAttr("DepartamentoEstado", n.Empleador.Departamento)
	emp.CreateAttr("MunicipioCiudad", n.Empleador.Municipio)
	emp.CreateAttr("Direccion", n.Empleador.Direccion)

	// Trabajador
	trab := root.CreateElement("Trabajador")
	trab.CreateAttr("TipoTrabajador", n.Trabajador.TipoTrabajador)
	trab.CreateAttr("SubTipoTrabajador", n.Trabajador.SubTipoTrabajador)
	trab.CreateAttr("AltoRiesgoPension", boolStr(n.Trabajador.AltoRiesgoPension))
	trab.CreateAttr("TipoDocumento", n.Trabajador.TipoDocumento)
	trab.CreateAttr("NumeroDocumento", n.Trabajador.NumeroDocumento)
	trab.CreateAttr("PrimerApellido", n.Trabajador.PrimerApellido)
	trab.CreateAttr("SegundoApellido", n.Trabajador.SegundoApellido)
	trab.CreateAttr("PrimerNombre", n.Trabajador.PrimerNombre)
	trab.CreateAttr("OtrosNombres", n.Trabajador.OtrosNombres)
	trab.CreateAttr("LugarTrabajoPais", n.Trabajador.LugarPais)
	trab.CreateAttr("LugarTrabajoDepartamentoEstado", n.Trabajador.LugarDepartamento)
	trab.CreateAttr("LugarTrabajoMunicipioCiudad", n.Trabajador.LugarMunicipio)
	trab.CreateAttr("LugarTrabajoDireccion", n.Trabajador.LugarDireccion)
	trab.CreateAttr("SalarioIntegral", boolStr(n.Trabajador.SalarioIntegral))
	trab.CreateAttr("TipoContrato", n.Trabajador.TipoContrato)
	trab.CreateAttr("Sueldo", domain.FormatCents(n.Trabajador.SueldoCents))
	trab.CreateAttr("CodigoTrabajador", n.Trabajador.CodigoTrabajador)

	// Pago
	pago := root.CreateElement("Pago")
	pago.CreateAttr("Forma", n.Pago.Forma)
	pago.CreateAttr("Metodo", n.Pago.Metodo)
	pago.CreateAttr("Banco", n.Pago.Banco)
	pago.CreateAttr("TipoCuenta", n.Pago.TipoCuenta)
	pago.CreateAttr("NumeroCuenta", n.Pago.NumeroCuenta)

	// FechasPagos
	if len(n.FechasPago) == 0 {
		return nil, fmt.Errorf("payroll: FechasPago must not be empty")
	}
	fechas := root.CreateElement("FechasPagos")
	for _, f := range n.FechasPago {
		fechas.CreateElement("FechaPago").SetText(f)
	}

	// Devengados
	dev := root.CreateElement("Devengados")
	basico := dev.CreateElement("Basico")
	basico.CreateAttr("DiasTrabajados", strconv.Itoa(n.Devengados.Basico.DiasTrabajados))
	basico.CreateAttr("SueldoTrabajado", domain.FormatCents(n.Devengados.Basico.SueldoTrabajadoCents))

	if t := n.Devengados.Transporte; t != nil {
		tr := dev.CreateElement("Transporte")
		tr.CreateAttr("AuxilioTransporte", domain.FormatCents(t.AuxilioTransporteCents))
		// ViaticoManuAlojS/NS are only emitted when > 0 (NIE072/073 rejects zeros).
		if t.ViaticoManuAlojSCents > 0 {
			tr.CreateAttr("ViaticoManuAlojS", domain.FormatCents(t.ViaticoManuAlojSCents))
		}
		if t.ViaticoManuAlojNSCents > 0 {
			tr.CreateAttr("ViaticoManuAlojNS", domain.FormatCents(t.ViaticoManuAlojNSCents))
		}
	}

	// Deducciones
	ded := root.CreateElement("Deducciones")
	if s := n.Deducciones.Salud; s != nil {
		sal := ded.CreateElement("Salud")
		sal.CreateAttr("Porcentaje", pct(s.Porcentaje))
		sal.CreateAttr("Deduccion", domain.FormatCents(s.DeduccionCents))
	}
	if fp := n.Deducciones.FondoPension; fp != nil {
		pension := ded.CreateElement("FondoPension")
		pension.CreateAttr("Porcentaje", pct(fp.Porcentaje))
		pension.CreateAttr("Deduccion", domain.FormatCents(fp.DeduccionCents))
	}
	if fsp := n.Deducciones.FondoSP; fsp != nil {
		fondoSP := ded.CreateElement("FondoSP")
		fondoSP.CreateAttr("Porcentaje", pct(fsp.Porcentaje))
		fondoSP.CreateAttr("DeduccionSP", domain.FormatCents(fsp.DeduccionSPCents))
		fondoSP.CreateAttr("PorcentajeSub", pct(fsp.PorcentajeSub))
		fondoSP.CreateAttr("DeduccionSub", domain.FormatCents(fsp.DeduccionSubCents))
	}
	if n.Deducciones.RetencionFuenteCents != 0 {
		ded.CreateElement("RetencionFuente").SetText(domain.FormatCents(n.Deducciones.RetencionFuenteCents))
	}

	// Totales
	root.CreateElement("Redondeo").SetText(domain.FormatCents(n.RedondeoCents))
	root.CreateElement("DevengadosTotal").SetText(domain.FormatCents(n.DevengadosTotalCents))
	root.CreateElement("DeduccionesTotal").SetText(domain.FormatCents(n.DeduccionesTotalCents))
	root.CreateElement("ComprobanteTotal").SetText(domain.FormatCents(n.ComprobanteTotalCents))

	return doc, nil
}

func pct(v float64) string { return fmt.Sprintf("%.2f", v) }
func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
