// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
package zip

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestBuild_RoundTrip(t *testing.T) {
	files := []File{
		{Name: "fv0800197268000190000000B.xml", Content: []byte("<Invoice/>")},
		{Name: "ar0800197268000190000000B.xml", Content: []byte("<ApplicationResponse/>")},
	}

	data, err := Build(files)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("abrir el zip generado: %v", err)
	}
	if len(r.File) != len(files) {
		t.Fatalf("se esperaban %d entradas, hay %d", len(files), len(r.File))
	}
	for i, want := range files {
		got := r.File[i]
		if got.Name != want.Name {
			t.Errorf("entrada %d: nombre = %q, want %q", i, got.Name, want.Name)
		}
		rc, err := got.Open()
		if err != nil {
			t.Fatalf("abrir entrada %s: %v", got.Name, err)
		}
		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("leer entrada %s: %v", got.Name, err)
		}
		rc.Close()
		if string(content) != string(want.Content) {
			t.Errorf("entrada %s: contenido = %q, want %q", got.Name, content, want.Content)
		}
	}
}

func TestDocumentFileName(t *testing.T) {
	cases := []struct {
		kind DocumentKind
		want string
	}{
		// NIT 800197268 (9 digits) -> 0800197268 (10 digits), in-house software, year 2019,
		// decimal sequence number 11 -> hex "0000000B" (not "00000011": see the note in
		// DocumentFileName about the inconsistency in the Annex's own illustrative example).
		{KindInvoice, "fv0800197268000190000000B.xml"},
		{KindCreditNote, "nc0800197268000190000000B.xml"},
		{KindDebitNote, "nd0800197268000190000000B.xml"},
		{KindSupportDocument, "ds0800197268000190000000B.xml"},
		{KindAdjustmentNote, "na0800197268000190000000B.xml"},
	}
	for _, c := range cases {
		got := DocumentFileName(c.kind, "800197268", SoftwarePropioCode, 2019, 11)
		if got != c.want {
			t.Errorf("DocumentFileName(%q) = %q, want %q", c.kind, got, c.want)
		}
	}
}

func TestPackageFileName(t *testing.T) {
	got := PackageFileName("800197268", SoftwarePropioCode, 2019, 11)
	want := "z0800197268000190000000B.zip"
	if got != want {
		t.Errorf("PackageFileName() = %q, want %q", got, want)
	}
}
