package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	xinvoice "github.com/andeedotnet/go-xinvoice"
)

// A complete, XRechnung-valid invoice: `validate` reports zero error findings.
const validJSON = `{
  "number": "R-2024-1",
  "issueDate": "2024-01-15",
  "dueDate": "2024-02-15",
  "typeCode": "380",
  "currencyCode": "EUR",
  "buyerReference": "04011000-12345-39",
  "businessProcessType": "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
  "specificationIdentifier": "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0",
  "seller": {"name": "Muster GmbH", "vatIdentifier": "DE123456789",
             "electronicAddress": {"value": "s@muster.de", "scheme": "EM"},
             "address": {"city": "Berlin", "postCode": "10115", "countryCode": "DE"},
             "contact": {"point": "Vertrieb", "phone": "+4930123", "email": "s@muster.de"}},
  "buyer": {"name": "Behörde", "electronicAddress": {"value": "b@amt.de", "scheme": "EM"},
            "address": {"city": "Hamburg", "postCode": "20095", "countryCode": "DE"}},
  "paymentInstructions": {"meansTypeCode": "58", "creditTransfers": [{"accountIdentifier": "DE02120300000000202051"}]},
  "totals": {"lineNetTotal": 100.00, "taxBasisTotal": 100.00, "taxTotal": 19.00,
             "grandTotal": 119.00, "duePayableAmount": 119.00},
  "vatBreakdown": [{"taxableAmount": 100.00, "taxAmount": 19.00, "categoryCode": "S", "rate": 19}],
  "lines": [{"id": "1", "quantity": {"value": 1, "unit": "C62"}, "netAmount": 100.00,
             "price": {"netPrice": 100.00}, "vat": {"categoryCode": "S", "rate": 19},
             "item": {"name": "Widget"}}]
}`

// A parseable but XRechnung-incomplete invoice (missing required fields), so
// `validate` must report error findings and exit non-zero.
const invalidJSON = `{
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

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

// captureStdout swaps os.Stdout for the duration of fn and returns what was
// written (the convert/validate commands print results there via fmt/os.Stdout).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

// TestConvertJSONToUBLToJSON exercises the model pivot through the CLI: JSON in,
// UBL out, then that UBL back to JSON (auto-detected). The invoice number must
// survive both hops.
func TestConvertJSONToUBLToJSON(t *testing.T) {
	in := writeTemp(t, "in.json", validJSON)
	ublPath := filepath.Join(t.TempDir(), "out_ubl.xml")
	if code := cmdConvert([]string{"--to", "ubl", "--in", in, "--out", ublPath}); code != 0 {
		t.Fatalf("convert --to ubl: exit %d, want 0", code)
	}
	ublXML, _ := os.ReadFile(ublPath)
	if !strings.Contains(string(ublXML), "Invoice") {
		t.Fatalf("UBL output does not look like UBL:\n%s", ublXML)
	}

	jsonPath := filepath.Join(t.TempDir(), "back.json")
	if code := cmdConvert([]string{"--to", "json", "--in", ublPath, "--out", jsonPath}); code != 0 {
		t.Fatalf("convert ubl --to json (auto-detect): exit %d, want 0", code)
	}
	back, _ := os.ReadFile(jsonPath)
	if !strings.Contains(string(back), `"R-2024-1"`) {
		t.Errorf("invoice number lost in JSON->UBL->JSON round-trip:\n%s", back)
	}
}

// TestConvertToCII checks the other syntax is wired and detectable.
func TestConvertToCII(t *testing.T) {
	in := writeTemp(t, "in.json", validJSON)
	out := captureStdout(t, func() {
		if code := cmdConvert([]string{"--to", "cii", "--in", in}); code != 0 {
			t.Errorf("convert --to cii: exit %d, want 0", code)
		}
	})
	if !strings.Contains(out, "CrossIndustryInvoice") {
		t.Errorf("CII output does not look like CII:\n%s", out)
	}
}

// TestConvertReadsStdin covers the default input path (--in "-").
func TestConvertReadsStdin(t *testing.T) {
	f, err := os.Open(writeTemp(t, "in.json", validJSON))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	old := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = old }()

	outPath := filepath.Join(t.TempDir(), "out.json")
	if code := cmdConvert([]string{"--to", "json", "--out", outPath}); code != 0 {
		t.Fatalf("convert from stdin: exit %d, want 0", code)
	}
	out, _ := os.ReadFile(outPath)
	if !strings.Contains(string(out), `"R-2024-1"`) {
		t.Errorf("stdin not read:\n%s", out)
	}
}

// TestConvertUnknownTarget rejects an unknown --to with usage exit code 2.
func TestConvertUnknownTarget(t *testing.T) {
	in := writeTemp(t, "in.json", validJSON)
	if code := cmdConvert([]string{"--to", "xml", "--in", in}); code != 2 {
		t.Errorf("convert --to xml: exit %d, want 2", code)
	}
}

// TestConvertMissingInput reports a read error (exit 1), not a panic.
func TestConvertMissingInput(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.json")
	if code := cmdConvert([]string{"--to", "json", "--in", missing}); code != 1 {
		t.Errorf("convert missing input: exit %d, want 1", code)
	}
}

// TestValidateExitCodes is the CI-relevant contract: 0 for a valid invoice, 1
// when there are error findings (so `validate` can gate a pipeline).
func TestValidateExitCodes(t *testing.T) {
	okPath := writeTemp(t, "ok.json", validJSON)
	var okCode int
	okOut := captureStdout(t, func() { okCode = cmdValidate([]string{"--in", okPath}) })
	if okCode != 0 {
		t.Errorf("validate valid: exit %d, want 0", okCode)
	}
	if !strings.Contains(okOut, `"valid": true`) {
		t.Errorf("valid invoice not reported valid:\n%s", okOut)
	}

	badPath := writeTemp(t, "bad.json", invalidJSON)
	var badCode int
	badOut := captureStdout(t, func() { badCode = cmdValidate([]string{"--in", badPath}) })
	if badCode != 1 {
		t.Errorf("validate invalid: exit %d, want 1", badCode)
	}
	if !strings.Contains(badOut, `"valid": false`) || !strings.Contains(badOut, `"rule"`) {
		t.Errorf("invalid invoice findings missing:\n%s", badOut)
	}
}

// TestValidateLanguage proves --lang flows through to localized messages: the
// German and English renderings of the same findings differ.
func TestValidateLanguage(t *testing.T) {
	bad := writeTemp(t, "bad.json", invalidJSON)
	de := captureStdout(t, func() { cmdValidate([]string{"--lang", "de", "--in", bad}) })
	en := captureStdout(t, func() { cmdValidate([]string{"--lang", "en", "--in", bad}) })
	if de == en {
		t.Error("--lang de and --lang en produced identical output; localization not wired")
	}
}

// TestParseInputAutoDetect confirms parseInput dispatches on content: '<' is XML,
// anything else is JSON.
func TestParseInputAutoDetect(t *testing.T) {
	jsonPath := writeTemp(t, "doc.json", validJSON)
	inv, err := parseInput(jsonPath)
	if err != nil {
		t.Fatalf("parseInput(json): %v", err)
	}
	ubl, err := inv.ToXML(xinvoice.UBL)
	if err != nil {
		t.Fatalf("ToXML: %v", err)
	}
	xmlPath := writeTemp(t, "doc.xml", string(ubl))
	if _, err := parseInput(xmlPath); err != nil {
		t.Fatalf("parseInput(xml): %v", err)
	}
}
