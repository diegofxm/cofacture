// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0
//
// Package zip compresses signed UBL documents into the package format required by DIAN's
// receiving web services (SendBillSync/SendBillAsync, Technical Annex 1.9, sections
// 6.5.7/6.5.8/7.8/7.10).
package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
)

// File is a document to include in the compressed package.
type File struct {
	Name    string
	Content []byte
}

// Build compresses the given documents into a single ZIP file, in the order received (a slice
// rather than a map: the order must be deterministic, not dependent on map iteration order).
//
// SendBillSync requires exactly one document; SendBillAsync allows up to 50 (section 6.5.8).
// Validating that count is the caller's responsibility — soap, not zip — because it depends on
// which operation will be used, something this package has no knowledge of.
func Build(files []File) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, f := range files {
		entry, err := w.Create(f.Name)
		if err != nil {
			return nil, fmt.Errorf("zip: create entry %s: %w", f.Name, err)
		}
		if _, err := entry.Write(f.Content); err != nil {
			return nil, fmt.Errorf("zip: write entry %s: %w", f.Name, err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("zip: close: %w", err)
	}
	return buf.Bytes(), nil
}
