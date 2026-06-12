package ubl

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/andeedotnet/go-xinvoice/model"
)

// TestSubInvoiceLineRoundTrip round-trips the official UBL extension instances
// that use nested sub-invoice lines (BG-DEX-01), proving the recursive mapping
// works. Uses the curated testdata/ or the full upstream bundle under knowledge/
// (https://xeinkauf.de/xrechnung/versionen-und-bundles/); skips when neither is
// present.
func TestSubInvoiceLineRoundTrip(t *testing.T) {
	root := findTestsuite(t)
	dir := filepath.Join(root, "extension")

	var withSubLines int
	for _, name := range []string{
		"04.01a-INVOICE_ubl.xml", "04.02a-INVOICE_ubl.xml",
		"04.03a-INVOICE_ubl.xml", "04.04a-INVOICE_ubl.xml",
	} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		t.Run(name, func(t *testing.T) {
			m1, err := Parse(raw)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if countSubLines(m1.Lines) == 0 {
				t.Error("expected sub-invoice lines, found none")
			} else {
				withSubLines++
			}
			again, err := Marshal(m1)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			m2, err := Parse(again)
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			if !reflect.DeepEqual(m1, m2) {
				j1, _ := m1.ToJSON()
				j2, _ := m2.ToJSON()
				t.Errorf("sub-line round-trip mismatch:\n%s", firstDiff(string(j1), string(j2)))
			}
		})
	}
	if withSubLines == 0 {
		t.Skip("no extension instances available")
	}
}

// TestThirdPartyPaymentRoundTrip round-trips the extension instance that uses
// cac:PrepaidPayment (third-party payments, BG-DEX-09).
func TestThirdPartyPaymentRoundTrip(t *testing.T) {
	root := findTestsuite(t)
	raw, err := os.ReadFile(filepath.Join(root, "extension", "05.01a-INVOICE_ubl.xml"))
	if err != nil {
		t.Skip("05.01a not available")
	}
	m1, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(m1.ThirdPartyPayments) == 0 {
		t.Error("expected third-party payments, found none")
	}
	again, err := Marshal(m1)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	m2, err := Parse(again)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if !reflect.DeepEqual(m1, m2) {
		j1, _ := m1.ToJSON()
		j2, _ := m2.ToJSON()
		t.Errorf("round-trip mismatch:\n%s", firstDiff(string(j1), string(j2)))
	}
}

func countSubLines(lines []model.Line) int {
	n := 0
	for i := range lines {
		n += len(lines[i].SubLines) + countSubLines(lines[i].SubLines)
	}
	return n
}
