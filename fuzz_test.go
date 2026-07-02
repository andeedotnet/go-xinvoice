package xinvoice_test

import (
	"os"
	"path/filepath"
	"testing"

	xinvoice "github.com/andeedotnet/go-xinvoice"
)

// xmlSeeds returns the committed testsuite instances as fuzz seeds (best-effort;
// empty when the testdata directory is absent, e.g. a trimmed checkout).
func xmlSeeds() [][]byte {
	var seeds [][]byte
	_ = filepath.WalkDir("testdata/instances", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".xml" {
			return nil
		}
		if b, e := os.ReadFile(path); e == nil {
			seeds = append(seeds, b)
		}
		return nil
	})
	return seeds
}

// FuzzParseXML asserts that ParseXML never panics on arbitrary input and that a
// successfully parsed model round-trips through serialization and validation
// without panicking. This guards the untrusted-input surface (parsers are the
// classic place for crashes and pathological-input DoS — see the Decimal-length
// guard in model/types.go).
func FuzzParseXML(f *testing.F) {
	f.Add([]byte(`<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"></rsm:CrossIndustryInvoice>`))
	f.Add([]byte(`<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"></Invoice>`))
	f.Add([]byte(``))
	for _, s := range xmlSeeds() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		inv, err := xinvoice.ParseXML(data)
		if err != nil || inv == nil {
			return
		}
		// A parsed model must serialize to both syntaxes, to JSON, and validate
		// without panicking. Errors are acceptable; panics are not.
		_, _ = inv.ToXML(xinvoice.UBL)
		_, _ = inv.ToXML(xinvoice.CII)
		_, _ = inv.ToJSON()
		_ = xinvoice.Validate(inv)
	})
}

// FuzzFromJSON asserts that FromJSON never panics on arbitrary input and that a
// successfully parsed model round-trips without panicking.
func FuzzFromJSON(f *testing.F) {
	f.Add([]byte(`{"number":"1","currencyCode":"EUR"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	// Derive realistic JSON seeds from the XML instances when available.
	for _, s := range xmlSeeds() {
		if inv, err := xinvoice.ParseXML(s); err == nil {
			if j, err := inv.ToJSON(); err == nil {
				f.Add(j)
			}
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		inv, err := xinvoice.FromJSON(data)
		if err != nil || inv == nil {
			return
		}
		_, _ = inv.ToJSON()
		_, _ = inv.ToXML(xinvoice.CII)
		_ = xinvoice.Validate(inv)
	})
}
