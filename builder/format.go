// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package builder

import (
	"fmt"

	"github.com/diegofxm/cofacture/domain"
)

// formatAmount renders cents as the 2-decimal string DIAN requires.
func formatAmount(cents int64) string {
	return domain.FormatCents(cents)
}

// formatQuantity renders a quantity with 6 decimals, as the technical annex requires.
func formatQuantity(q float64) string {
	return fmt.Sprintf("%.6f", q)
}

// formatPercent renders a tax percentage with 2 decimals.
func formatPercent(p float64) string {
	return fmt.Sprintf("%.2f", p)
}
