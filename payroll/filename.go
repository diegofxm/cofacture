// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package payroll

import "fmt"

// XMLFileName builds a NominaIndividual's XML file name:
//
//	nie{NIT, 10 digits, left-padded}{year, 2 digits}{consecutive, 8 hex}.xml
//
// Section 3.3 of the Electronic Payroll Technical Annex. The consecutive resets on January 1st
// each year; keeping track of it is the orchestrator's responsibility.
func XMLFileName(nit string, year int, consecutive uint32) string {
	return fmt.Sprintf("nie%010s%02d%08X.xml", nit, year%100, consecutive)
}

// AdjustXMLFileName builds a NominaIndividualDeAjuste's XML file name:
//
//	niae{NIT, 10 digits}{year, 2 digits}{consecutive, 8 hex}.xml
func AdjustXMLFileName(nit string, year int, consecutive uint32) string {
	return fmt.Sprintf("niae%010s%02d%08X.xml", nit, year%100, consecutive)
}

// ZIPFileName builds the ZIP file name that contains the payroll document:
//
//	z{NIT, 10 digits}{year, 2 digits}{consecutive, 8 hex}.zip
//
// Section 3.5 of the Technical Annex. This consecutive belongs to the sent ZIP package,
// distinct from the individual XML files' consecutive.
func ZIPFileName(nit string, year int, consecutive uint32) string {
	return fmt.Sprintf("z%010s%02d%08X.zip", nit, year%100, consecutive)
}
