// Command gen builds the embedded catalog (internal/catalog/*_gen.go) from the
// official KoSIT XRechnung bundle (downloaded from
// https://xeinkauf.de/xrechnung/versionen-und-bundles/).
//
// It is run via `go generate ./internal/catalog`. Sources and what is taken
// from each:
//
//   - SeMoX model XML        -> German rule texts + the BT/BG terms each rule is about
//   - EN16931 *.xsl          -> English EN16931 core rule texts + severity, and the
//     BR-CL-* enumerated code lists
//   - XRechnung *.sch         -> BR-DE (German) and PEPPOL (English) rule texts + severity
//   - visualization l10n xml -> German/English BT/BG labels
//
// Output is gofmt-formatted and deterministic (entries sorted by id).
package main

import (
	"bytes"
	"embed"
	"encoding/xml"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

// resources holds verbatim copies of the KoSIT source artifacts the catalog is
// generated from. They are embedded so the generator runs without the full
// upstream bundle present. To refresh them after a bundle update, re-download the
// bundle (https://xeinkauf.de/xrechnung/versionen-und-bundles/), re-copy the
// files described in resources/README.md and re-run `go generate ./...`.
//
//go:embed resources
var resources embed.FS

func main() {
	log.SetFlags(0)

	rules := map[string]*rule{}
	collectModelRules(rules)      // German text + terms
	collectXSLRules(rules)        // English EN16931 text + severity
	collectSchematronRules(rules) // BR-DE/PEPPOL text + severity
	labels := collectLabels()
	codeLists := collectCodeLists()

	writeRules("messages_gen.go", rules)
	writeLabels("labels_gen.go", labels)
	writeCodeLists("codelists_gen.go", codeLists)

	log.Printf("gen: wrote %d rules, %d labels, %d code lists", len(rules), len(labels), len(codeLists))
}

// ---------------------------------------------------------------------------
// Data carried while generating
// ---------------------------------------------------------------------------

type rule struct {
	ID       string
	Severity string
	DE       string
	EN       string
	Terms    []string
}

type label struct {
	ID string
	DE string
	EN string
}

// ---------------------------------------------------------------------------
// Source 1: SeMoX model XML — German rule texts + terms
// ---------------------------------------------------------------------------

func collectModelRules(out map[string]*rule) {
	path := "resources/xrechnung-cius-model.xml"
	d := xml.NewDecoder(mustOpen(path))
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("gen: model %s: %v", path, err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "rule" {
			continue
		}
		id := attr(se, "id")
		if id == "" {
			d.Skip()
			continue
		}
		r := get(out, id)
		r.Terms = splitTerms(attr(se, "on-terms"))
		// Find the <description> child and take its ordered text.
		for {
			t, err := d.Token()
			if err != nil {
				break
			}
			if s, ok := t.(xml.StartElement); ok && s.Name.Local == "description" {
				r.DE = normalizeWS(text(d))
			}
			if e, ok := t.(xml.EndElement); ok && e.Name.Local == "rule" {
				break
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Source 2: EN16931 compiled XSL — English core rule texts + severity
// ---------------------------------------------------------------------------

var bracketPrefix = regexp.MustCompile(`^\s*\[[^\]]*\]\s*-?\s*`)

func collectXSLRules(out map[string]*rule) {
	for _, path := range []string{
		"resources/EN16931-UBL-validation.xsl",
		"resources/EN16931-CII-validation.xsl",
	} {
		d := xml.NewDecoder(mustOpen(path))
		var cur *rule // failed-assert currently open
		for {
			tok, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("gen: xsl %s: %v", path, err)
			}
			switch se := tok.(type) {
			case xml.StartElement:
				switch se.Name.Local {
				case "failed-assert":
					id := attr(se, "id")
					if id == "" || isSyntaxRule(id) {
						cur = nil
						continue
					}
					cur = get(out, id)
					if cur.Severity == "" {
						cur.Severity = severity(attr(se, "flag"))
					}
				case "text":
					if cur != nil && cur.EN == "" {
						cur.EN = stripRulePrefix(normalizeWS(text(d)))
					}
				}
			case xml.EndElement:
				if se.Name.Local == "failed-assert" {
					cur = nil
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Source 3: XRechnung schematron — BR-DE (German) / PEPPOL (English) + severity
// ---------------------------------------------------------------------------

func collectSchematronRules(out map[string]*rule) {
	for _, path := range []string{
		"resources/XRechnung-UBL-validation.sch",
		"resources/XRechnung-CII-validation.sch",
	} {
		d := xml.NewDecoder(mustOpen(path))
		for {
			tok, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("gen: sch %s: %v", path, err)
			}
			se, ok := tok.(xml.StartElement)
			if !ok || se.Name.Local != "assert" {
				continue
			}
			id := attr(se, "id")
			if id == "" {
				continue
			}
			r := get(out, id)
			if r.Severity == "" {
				r.Severity = severity(attr(se, "flag"))
			}
			txt := stripRulePrefix(normalizeWS(text(d)))
			// BR-DE national rules are written in German, PEPPOL rules in English.
			if strings.HasPrefix(id, "BR-DE") {
				if r.DE == "" {
					r.DE = txt
				}
			} else if r.EN == "" {
				r.EN = txt
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Source 4: visualization l10n — BT/BG labels
// ---------------------------------------------------------------------------

func collectLabels() map[string]*label {
	out := map[string]*label{}
	load := func(path, lang string) {
		d := xml.NewDecoder(mustOpen(path))
		for {
			tok, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("gen: l10n %s: %v", path, err)
			}
			se, ok := tok.(xml.StartElement)
			if !ok || se.Name.Local != "entry" {
				continue
			}
			id := attr(se, "id")
			val := normalizeWS(text(d))
			if id == "" {
				continue
			}
			l := out[id]
			if l == nil {
				l = &label{ID: id}
				out[id] = l
			}
			if lang == "de" {
				l.DE = val
			} else {
				l.EN = val
			}
		}
	}
	load("resources/l10n-de.xml", "de")
	load("resources/l10n-en.xml", "en")
	return out
}

// ---------------------------------------------------------------------------
// Source 5: EN16931 XSL — BR-CL-* enumerated code lists
// ---------------------------------------------------------------------------

// clTestRe captures the test expression of each BR-CL failed-assert together
// with its rule id. The test is element text (so `<`/`>` are escaped) — the only
// literal `<` is the closing tag, hence [^<]* matches the whole expression and
// cannot leak across rules.
var clTestRe = regexp.MustCompile(`(?s)<xsl:attribute name="test">([^<]*)</xsl:attribute>\s*<svrl:text>\[(BR-CL[0-9A-Za-z-]*)\]`)

// clContainsRe captures the enumerated list literal of a `contains(' a b c ', …)`
// membership test (the schematron's own code-list mechanism).
var clContainsRe = regexp.MustCompile(`contains\(\s*'([^']*)'`)

// collectCodeLists extracts, per BR-CL rule, the union of values enumerated in
// the EN16931 XSL membership tests. BR-CL-24 (MIME) uses a regex rather than an
// enumeration and so yields no list.
func collectCodeLists() map[string]string {
	sets := map[string]map[string]bool{}
	for _, name := range []string{
		"resources/EN16931-UBL-validation.xsl",
		"resources/EN16931-CII-validation.xsl",
	} {
		data, err := resources.ReadFile(name)
		if err != nil {
			log.Fatalf("gen: open embedded %s: %v", name, err)
		}
		for _, m := range clTestRe.FindAllSubmatch(data, -1) {
			rid := canonicalID(string(m[2]))
			set := sets[rid]
			if set == nil {
				set = map[string]bool{}
				sets[rid] = set
			}
			for _, c := range clContainsRe.FindAllSubmatch(m[1], -1) {
				for _, tok := range strings.Fields(string(c[1])) {
					set[tok] = true
				}
			}
		}
	}
	// XRechnung profiles extend some EN16931 base code lists. Add those values so
	// the lists match what XRechnung actually accepts.
	xrExtensions := map[string][]string{
		"BR-CL-13": {"CVD"}, // CVD construction profile item classification scheme
	}
	for rid, extra := range xrExtensions {
		set := sets[canonicalID(rid)]
		if set == nil {
			set = map[string]bool{}
			sets[canonicalID(rid)] = set
		}
		for _, v := range extra {
			set[v] = true
		}
	}

	out := map[string]string{}
	for rid, set := range sets {
		if len(set) == 0 {
			continue
		}
		vals := make([]string, 0, len(set))
		for v := range set {
			vals = append(vals, v)
		}
		sort.Strings(vals)
		out[rid] = " " + strings.Join(vals, " ") + " "
	}
	return out
}

// ---------------------------------------------------------------------------
// Emitting Go source
// ---------------------------------------------------------------------------

const header = "// Code generated by internal/gen; DO NOT EDIT.\n\npackage catalog\n\n"

func writeRules(name string, rules map[string]*rule) {
	ids := make([]string, 0, len(rules))
	for id := range rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b bytes.Buffer
	b.WriteString(header)
	fmt.Fprintf(&b, "// rules holds %d EN16931 / XRechnung business rules keyed by rule id.\n", len(ids))
	b.WriteString("var rules = map[string]Rule{\n")
	for _, id := range ids {
		r := rules[id]
		fmt.Fprintf(&b, "\t%q: {ID: %q, Severity: %s, DE: %q, EN: %q", id, id, severityConst(r.Severity), r.DE, r.EN)
		if len(r.Terms) > 0 {
			b.WriteString(", Terms: []string{")
			for i, t := range r.Terms {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%q", t)
			}
			b.WriteString("}")
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n")
	writeFormatted(name, b.Bytes())
}

func writeLabels(name string, labels map[string]*label) {
	ids := make([]string, 0, len(labels))
	for id := range labels {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b bytes.Buffer
	b.WriteString(header)
	fmt.Fprintf(&b, "// labels holds %d BT/BG term labels keyed by term id.\n", len(ids))
	b.WriteString("var labels = map[string]Label{\n")
	for _, id := range ids {
		l := labels[id]
		fmt.Fprintf(&b, "\t%q: {ID: %q, DE: %q, EN: %q},\n", id, id, l.DE, l.EN)
	}
	b.WriteString("}\n")
	writeFormatted(name, b.Bytes())
}

func writeCodeLists(name string, lists map[string]string) {
	ids := make([]string, 0, len(lists))
	for id := range lists {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b bytes.Buffer
	b.WriteString(header)
	fmt.Fprintf(&b, "// codeLists holds the enumerated EN16931 code-list values for %d BR-CL rules,\n", len(ids))
	b.WriteString("// space-delimited with leading/trailing spaces (matching the schematron membership\n")
	b.WriteString("// test: a value v is in the list iff the string contains \" \"+v+\" \").\n")
	b.WriteString("var codeLists = map[string]string{\n")
	for _, id := range ids {
		fmt.Fprintf(&b, "\t%q: %q,\n", id, lists[id])
	}
	b.WriteString("}\n")
	writeFormatted(name, b.Bytes())
}

func writeFormatted(name string, src []byte) {
	formatted, err := format.Source(src)
	if err != nil {
		log.Fatalf("gen: format %s: %v\n%s", name, err, src)
	}
	if err := os.WriteFile(name, formatted, 0o644); err != nil {
		log.Fatalf("gen: write %s: %v", name, err)
	}
}

func severityConst(s string) string {
	switch s {
	case "error":
		return "SeverityError"
	case "warning":
		return "SeverityWarning"
	case "information":
		return "SeverityInfo"
	default:
		return "SeverityError"
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func get(m map[string]*rule, id string) *rule {
	id = canonicalID(id)
	r := m[id]
	if r == nil {
		r = &rule{ID: id}
		m[id] = r
	}
	return r
}

// canonicalID normalizes the zero-padding mismatch between sources: the SeMoX
// model writes "BR-1" / "BR-S-8" while the EN16931 XSL writes "BR-01" /
// "BR-S-08". Leading zeros are stripped from purely numeric "-NNN" segments,
// so both map to the same key. Segments that are not all digits (e.g. the
// "R001" of "PEPPOL-EN16931-R001", or the "a"/"b" suffixes) are left untouched.
func canonicalID(id string) string {
	parts := strings.Split(id, "-")
	for i, p := range parts {
		if isAllDigits(p) {
			trimmed := strings.TrimLeft(p, "0")
			if trimmed == "" {
				trimmed = "0"
			}
			parts[i] = trimmed
		}
	}
	return strings.Join(parts, "-")
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func mustOpen(name string) io.Reader {
	f, err := resources.ReadFile(name)
	if err != nil {
		log.Fatalf("gen: open embedded %s: %v", name, err)
	}
	return bytes.NewReader(f)
}

func attr(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// text consumes tokens until the end of the element whose start was just read,
// returning the concatenation of all character data in document order.
func text(d *xml.Decoder) string {
	var b strings.Builder
	depth := 0
	for {
		tok, err := d.Token()
		if err != nil {
			return b.String()
		}
		switch t := tok.(type) {
		case xml.CharData:
			b.Write(t)
		case xml.StartElement:
			depth++
		case xml.EndElement:
			if depth == 0 {
				return b.String()
			}
			depth--
		}
	}
}

var wsRe = regexp.MustCompile(`\s+`)

func normalizeWS(s string) string {
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

func stripRulePrefix(s string) string {
	return strings.TrimSpace(bracketPrefix.ReplaceAllString(s, ""))
}

func severity(flag string) string {
	switch strings.ToLower(flag) {
	case "fatal", "error", "":
		if flag == "" {
			return ""
		}
		return "error"
	case "warning":
		return "warning"
	case "information", "info":
		return "information"
	default:
		return "error"
	}
}

// isSyntaxRule reports whether id is a syntax-binding rule (UBL-CR/SR/DT,
// CII-SR/DT) rather than a semantic EN16931/XRechnung business rule. Those are
// enforced structurally by the typed UBL/CII parsers, so they are not carried
// in the semantic rule catalog.
func isSyntaxRule(id string) bool {
	return strings.HasPrefix(id, "UBL-") || strings.HasPrefix(id, "CII-")
}

func splitTerms(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' })
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
