package validation

import (
	"bytes"
	"testing"

	"github.com/andeedotnet/go-xinvoice/model"
)

// TestJSONIncludesLocationLabel checks that findings carry the BT/BG term label
// for their location, in the requested language.
func TestJSONIncludesLocationLabel(t *testing.T) {
	b, err := Validate(&model.Invoice{}).JSON("de")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"locationLabel"`)) {
		t.Fatal("findings should include a locationLabel field")
	}
	// An empty invoice violates BR-2 (location BT-1, label "Rechnungsnummer").
	if !bytes.Contains(b, []byte("Rechnungsnummer")) {
		t.Errorf("expected the BT-1 label (Rechnungsnummer) in the German output")
	}

	en, _ := Validate(&model.Invoice{}).JSON("en")
	if !bytes.Contains(en, []byte("Invoice number")) {
		t.Errorf("expected the BT-1 label (Invoice number) in the English output")
	}
}
