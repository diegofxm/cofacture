// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package domain

import "testing"

// TestFormatCents locks in the exact 2-decimal, no-thousands-separator format DIAN requires —
// this function is shared by every XML serializer and every CUFE/CUDE/CUDS hash composer in the
// library, so a regression here would silently corrupt both at once.
func TestFormatCents(t *testing.T) {
	cases := []struct {
		cents int64
		want  string
	}{
		{0, "0.00"},
		{5, "0.05"},
		{100, "1.00"},
		{350_000_000, "3500000.00"},
		{-500, "-5.00"},
		{-5, "-0.05"},
	}

	for _, c := range cases {
		if got := FormatCents(c.cents); got != c.want {
			t.Errorf("FormatCents(%d) = %q, want %q", c.cents, got, c.want)
		}
	}
}
