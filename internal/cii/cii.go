package cii

import (
	"bytes"
	"encoding/xml"

	"github.com/andeedotnet/go-xinvoice/internal/xmlfmt"
	"github.com/andeedotnet/go-xinvoice/model"
)

// ciiPrefixes maps the CII namespaces to their conventional prefixes for output.
var ciiPrefixes = map[string]string{
	nsRSM: "rsm",
	nsRAM: "ram",
	nsUDT: "udt",
	nsQDT: "qdt",
}

// Detect reports whether b looks like a CII document (root element
// CrossIndustryInvoice in the rsm namespace).
func Detect(b []byte) bool {
	d := xml.NewDecoder(bytes.NewReader(b))
	for {
		tok, err := d.Token()
		if err != nil {
			return false
		}
		if se, ok := tok.(xml.StartElement); ok {
			return se.Name.Space == nsRSM && se.Name.Local == "CrossIndustryInvoice"
		}
	}
}

// Parse unmarshals CII XML into the syntax-neutral model.
func Parse(b []byte) (*model.Invoice, error) {
	var c Invoice
	if err := xml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return toModel(&c), nil
}

// Marshal serializes the model as a CII CrossIndustryInvoice document with
// conventional rsm:/ram:/udt:/qdt: prefixes.
func Marshal(inv *model.Invoice) ([]byte, error) {
	c := fromModel(inv)
	body, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, err
	}
	prefixed, err := xmlfmt.Prefix(body, ciiPrefixes)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.Write(prefixed)
	return buf.Bytes(), nil
}
