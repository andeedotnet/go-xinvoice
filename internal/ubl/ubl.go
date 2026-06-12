package ubl

import (
	"bytes"
	"encoding/xml"

	"github.com/andeedotnet/go-xinvoice/internal/xmlfmt"
	"github.com/andeedotnet/go-xinvoice/model"
)

// ublPrefixes maps the UBL namespaces to their conventional prefixes for output.
var ublPrefixes = map[string]string{
	nsInvoice:    "ubl",
	nsCreditNote: "ubl",
	nsCBC:        "cbc",
	nsCAC:        "cac",
}

// Detect reports whether b looks like a UBL Invoice or CreditNote document.
func Detect(b []byte) bool {
	d := xml.NewDecoder(bytes.NewReader(b))
	for {
		tok, err := d.Token()
		if err != nil {
			return false
		}
		if se, ok := tok.(xml.StartElement); ok {
			return (se.Name.Space == nsInvoice && se.Name.Local == "Invoice") ||
				(se.Name.Space == nsCreditNote && se.Name.Local == "CreditNote")
		}
	}
}

// Parse unmarshals UBL Invoice XML into the syntax-neutral model.
func Parse(b []byte) (*model.Invoice, error) {
	var u Invoice
	if err := xml.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return toModel(&u), nil
}

// Marshal serializes the model as a UBL Invoice (or CreditNote) document with
// conventional ubl:/cbc:/cac: prefixes.
func Marshal(inv *model.Invoice) ([]byte, error) {
	u := fromModel(inv)
	body, err := xml.MarshalIndent(u, "", "  ")
	if err != nil {
		return nil, err
	}
	prefixed, err := xmlfmt.Prefix(body, ublPrefixes)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.Write(prefixed)
	return buf.Bytes(), nil
}
