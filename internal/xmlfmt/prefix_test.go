package xmlfmt

import (
	"encoding/xml"
	"strings"
	"testing"
)

const (
	nsInv = "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
	nsCBC = "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
	nsCAC = "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
)

func TestPrefix(t *testing.T) {
	verbose := []byte(`<Invoice xmlns="` + nsInv + `">` +
		`<ID xmlns="` + nsCBC + `">R&amp;1</ID>` +
		`<Line xmlns="` + nsCAC + `"><Amount xmlns="` + nsCBC + `" currencyID="EUR">9.50</Amount></Line>` +
		`</Invoice>`)

	out, err := Prefix(verbose, map[string]string{
		nsInv: "ubl", nsCBC: "cbc", nsCAC: "cac",
		"urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2": "ubl", // unused, must be ignored
	})
	if err != nil {
		t.Fatalf("Prefix: %v", err)
	}
	s := string(out)

	for _, want := range []string{
		`<ubl:Invoice`, `xmlns:cbc="` + nsCBC + `"`, `xmlns:cac="` + nsCAC + `"`,
		`<cbc:ID>R&amp;1</cbc:ID>`, `<cac:Line>`, `<cbc:Amount currencyID="EUR">9.50</cbc:Amount>`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q\n%s", want, s)
		}
	}
	// The unused CreditNote namespace must not be declared.
	if strings.Contains(s, "CreditNote-2") {
		t.Error("declared an unused namespace")
	}

	// The result must be well-formed and preserve namespaces.
	var n struct {
		XMLName xml.Name
		ID      string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID"`
	}
	if err := xml.Unmarshal(out, &n); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if n.XMLName.Space != nsInv || n.ID != "R&1" {
		t.Errorf("namespaces/values not preserved: %+v", n)
	}
}

func TestPrefixUnknownNamespace(t *testing.T) {
	verbose := []byte(`<X xmlns="urn:unknown"/>`)
	if _, err := Prefix(verbose, map[string]string{nsInv: "ubl"}); err == nil {
		t.Error("expected an error for an unmapped namespace")
	}
}
