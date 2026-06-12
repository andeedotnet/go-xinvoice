package model

import (
	"bytes"
	"encoding/json"
)

// ToJSON serializes the invoice to indented JSON following the semantic model.
func (inv *Invoice) ToJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(inv); err != nil {
		return nil, err
	}
	// json.Encoder appends a trailing newline; trim it for a clean document.
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// FromJSON parses a semantic-model JSON document into an [Invoice]. Unknown
// fields are ignored so that documents written by newer minor versions still load.
func FromJSON(b []byte) (*Invoice, error) {
	var inv Invoice
	if err := json.Unmarshal(b, &inv); err != nil {
		return nil, err
	}
	return &inv, nil
}
