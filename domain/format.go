// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

import "fmt"

// FormatCents renders cents as the 2-decimal string DIAN requires (no thousands separator,
// dot as the decimal separator) — used both when serializing XML and when composing the
// CUFE/CUDE hash input strings, which share the same format.
func FormatCents(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	s := fmt.Sprintf("%d.%02d", cents/100, cents%100)
	if neg {
		s = "-" + s
	}
	return s
}
