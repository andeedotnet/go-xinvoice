package xinvoice_test

import (
	"errors"
	"strings"
	"testing"

	xinvoice "github.com/andeedotnet/go-xinvoice"
)

const minimalJSON = `{
  "number": "R1",
  "issueDate": "2024-01-01",
  "typeCode": "380",
  "currencyCode": "EUR",
  "seller": {"name": "S", "address": {"countryCode": "DE"}},
  "buyer": {"name": "B", "address": {"countryCode": "DE"}},
  "totals": {"lineNetTotal": 10.00, "taxBasisTotal": 10.00, "grandTotal": 11.90, "duePayableAmount": 11.90},
  "vatBreakdown": [{"taxableAmount": 10.00, "taxAmount": 1.90, "categoryCode": "S", "rate": 19}],
  "lines": [{"id": "1", "quantity": {"value": 1, "unit": "C62"}, "netAmount": 10.00,
             "price": {"netPrice": 10.00}, "vat": {"categoryCode": "S", "rate": 19},
             "item": {"name": "X"}}]
}`

func TestFacadeJSONRoundTrip(t *testing.T) {
	inv, err := xinvoice.FromJSON([]byte(minimalJSON))
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}
	if inv.Number != "R1" || inv.Lines[0].Quantity.Value != "1" {
		t.Fatalf("unexpected parse: %+v", inv)
	}

	out, err := inv.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	again, err := xinvoice.FromJSON(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if again.Totals.GrandTotal != "11.90" {
		t.Errorf("grand total lost in round-trip: %q", again.Totals.GrandTotal)
	}
}

// Both UBL (P3) and CII (P4) are wired into the facade with auto-detection.
func TestFacadeXMLWiring(t *testing.T) {
	inv, _ := xinvoice.FromJSON([]byte(minimalJSON))

	for _, syntax := range []struct {
		name string
		s    xinvoice.Syntax
	}{{"UBL", xinvoice.UBL}, {"CII", xinvoice.CII}} {
		xmlBytes, err := inv.ToXML(syntax.s)
		if err != nil {
			t.Fatalf("ToXML(%s): %v", syntax.name, err)
		}
		back, err := xinvoice.ParseXML(xmlBytes) // auto-detects the syntax
		if err != nil {
			t.Fatalf("ParseXML(%s): %v", syntax.name, err)
		}
		if back.Number != "R1" || back.Lines[0].Item.Name != "X" {
			t.Errorf("%s facade round-trip lost data: %+v", syntax.name, back)
		}
	}

	// An unrecognized document is reported, not parsed.
	if _, err := xinvoice.ParseXML([]byte(`<html/>`)); !errors.Is(err, xinvoice.ErrNotImplemented) {
		t.Errorf("ParseXML(unknown) err = %v, want ErrNotImplemented", err)
	}
}

func TestFacadeValidate(t *testing.T) {
	inv, _ := xinvoice.FromJSON([]byte(minimalJSON))

	// The minimal document is missing XRechnung-required fields, so validation
	// must flag it and render localized findings.
	res := xinvoice.Validate(inv)
	if res.Valid() {
		t.Error("minimal invoice unexpectedly reported valid")
	}
	de, err := res.JSON("de")
	if err != nil {
		t.Fatalf("JSON(de): %v", err)
	}
	if !strings.Contains(string(de), `"valid": false`) || !strings.Contains(string(de), `"rule"`) {
		t.Errorf("unexpected findings JSON: %s", de)
	}

	// ValidateXML wires parsing + validation together (here via UBL).
	ublBytes, _ := inv.ToXML(xinvoice.UBL)
	if xinvoice.ValidateXML(ublBytes) == nil {
		t.Error("ValidateXML returned nil")
	}
}
