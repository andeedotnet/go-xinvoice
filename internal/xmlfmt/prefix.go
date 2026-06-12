// Package xmlfmt rewrites the verbose, per-element-xmlns output of
// encoding/xml into conventional prefixed XML (e.g. <ubl:Invoice><cbc:ID>…),
// with namespace declarations only on the root element.
//
// It is a deterministic post-processor: it decodes the verbose bytes (so each
// element's namespace is resolved) and re-emits prefixed, indented XML. The
// input must be non-mixed-content (every element has either child elements or
// text, never both), which is true for UBL and CII invoices.
package xmlfmt

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// Prefix rewrites verbose into prefixed form using nsToPrefix (namespace URI →
// prefix). Only namespaces actually present in the document are declared, so a
// map may list more namespaces than appear (e.g. both the Invoice and CreditNote
// roots).
func Prefix(verbose []byte, nsToPrefix map[string]string) ([]byte, error) {
	used, err := usedNamespaces(verbose)
	if err != nil {
		return nil, err
	}
	for ns := range used {
		if _, ok := nsToPrefix[ns]; !ok {
			return nil, fmt.Errorf("xmlfmt: no prefix configured for namespace %q", ns)
		}
	}

	dec := xml.NewDecoder(bytes.NewReader(verbose))
	var b bytes.Buffer
	depth := 0
	first := true

	type frame struct {
		qname    string
		hasChild bool
	}
	var stack []frame
	var text string // character data accumulated in the innermost open element

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			qname := nsToPrefix[t.Name.Space] + ":" + t.Name.Local
			if n := len(stack); n > 0 {
				stack[n-1].hasChild = true
			}
			if !first {
				b.WriteByte('\n')
				b.WriteString(strings.Repeat("  ", depth))
			}
			first = false
			b.WriteByte('<')
			b.WriteString(qname)
			if depth == 0 {
				writeRootNamespaces(&b, used, nsToPrefix)
			}
			writeAttrs(&b, t.Attr)
			b.WriteByte('>')
			stack = append(stack, frame{qname: qname})
			depth++
			text = ""
		case xml.CharData:
			text += string(t)
		case xml.EndElement:
			depth--
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch {
			case f.hasChild:
				b.WriteByte('\n')
				b.WriteString(strings.Repeat("  ", depth))
				b.WriteString("</" + f.qname + ">")
			case strings.TrimSpace(text) != "":
				_ = xml.EscapeText(&b, []byte(text))
				b.WriteString("</" + f.qname + ">")
			default:
				b.WriteString("</" + f.qname + ">")
			}
			text = ""
		}
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func usedNamespaces(verbose []byte) (map[string]bool, error) {
	dec := xml.NewDecoder(bytes.NewReader(verbose))
	used := map[string]bool{}
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		if se, ok := tok.(xml.StartElement); ok {
			used[se.Name.Space] = true
		}
	}
	return used, nil
}

// writeRootNamespaces declares xmlns:prefix for every used namespace, ordered by
// prefix for deterministic output.
func writeRootNamespaces(b *bytes.Buffer, used map[string]bool, nsToPrefix map[string]string) {
	type decl struct{ prefix, ns string }
	var decls []decl
	for ns := range used {
		decls = append(decls, decl{nsToPrefix[ns], ns})
	}
	sort.Slice(decls, func(i, j int) bool { return decls[i].prefix < decls[j].prefix })
	for _, d := range decls {
		b.WriteString(` xmlns:` + d.prefix + `="` + d.ns + `"`)
	}
}

// writeAttrs emits non-namespace attributes (currencyID, unitCode, schemeID, …).
func writeAttrs(b *bytes.Buffer, attrs []xml.Attr) {
	for _, a := range attrs {
		if a.Name.Local == "xmlns" || a.Name.Space == "xmlns" {
			continue
		}
		name := a.Name.Local
		if a.Name.Space != "" {
			name = a.Name.Space + ":" + a.Name.Local
		}
		b.WriteString(" " + name + `="`)
		_ = xml.EscapeText(b, []byte(a.Value))
		b.WriteByte('"')
	}
}
