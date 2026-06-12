// Package catalog holds reference data derived from the official KoSIT
// XRechnung bundle: the EN16931 / XRechnung business rules (with German and
// English texts and severity), the BT/BG term labels, and the BR-CL-* code lists.
//
// The data in *_gen.go is generated; regenerate it with `go generate ./...`
// after refreshing the source copies in internal/gen/resources/. See internal/gen.
package catalog

import (
	"sort"
	"strings"
)

// RuleIDs returns every rule id in the catalog, sorted.
func RuleIDs() []string {
	ids := make([]string, 0, len(rules))
	for id := range rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

//go:generate go run ../gen

// Severity classifies how a violated rule is reported, mirroring the schematron
// flag (fatal -> error).
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "information"
)

// Rule is one EN16931 / XRechnung business rule. DE comes from the SeMoX model,
// EN from the EN16931 schematron/XSL (and PEPPOL rules); Terms lists the BT/BG
// identifiers the rule constrains (may be empty).
type Rule struct {
	ID       string
	Severity Severity
	DE       string
	EN       string
	Terms    []string
}

// Text returns the rule message in the requested language ("de"/"en"), falling
// back to the other language when the preferred one is unavailable.
func (r Rule) Text(lang string) string {
	if lang == "de" {
		if r.DE != "" {
			return r.DE
		}
		return r.EN
	}
	if r.EN != "" {
		return r.EN
	}
	return r.DE
}

// Label is a localized BT/BG term label.
type Label struct {
	ID string
	DE string
	EN string
}

// Text returns the label in the requested language, with fallback.
func (l Label) Text(lang string) string {
	if lang == "de" && l.DE != "" {
		return l.DE
	}
	if lang == "en" && l.EN != "" {
		return l.EN
	}
	if l.EN != "" {
		return l.EN
	}
	return l.DE
}

// canon normalizes a rule/term id by stripping leading zeros from purely numeric
// "-NNN" segments, so callers may pass either the padded EN16931 form
// ("BR-CL-04") or the model form ("BR-CL-4"). It mirrors the generator's id
// canonicalization, which is the form actually stored in the maps.
func canon(id string) string {
	parts := strings.Split(id, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		allDigit := true
		for _, c := range p {
			if c < '0' || c > '9' {
				allDigit = false
				break
			}
		}
		if allDigit {
			t := strings.TrimLeft(p, "0")
			if t == "" {
				t = "0"
			}
			parts[i] = t
		}
	}
	return strings.Join(parts, "-")
}

// LookupRule returns the rule with the given id (padded or unpadded).
func LookupRule(id string) (Rule, bool) {
	r, ok := rules[canon(id)]
	return r, ok
}

// LookupLabel returns the label for the given BT/BG term id.
func LookupLabel(id string) (Label, bool) {
	l, ok := labels[canon(id)]
	return l, ok
}

// Rules returns the number of rules in the catalog.
func Rules() int { return len(rules) }

// Labels returns the number of labels in the catalog.
func Labels() int { return len(labels) }

// InCodeList reports whether value is a member of the code list enumerated by
// the given BR-CL rule. The second result is false when no list is known for the
// rule (e.g. BR-CL-24, whose MIME check is a regex rather than an enumeration).
func InCodeList(ruleID, value string) (inList, known bool) {
	list, ok := codeLists[canon(ruleID)]
	if !ok {
		return false, false
	}
	return strings.Contains(list, " "+strings.TrimSpace(value)+" "), true
}

// CodeListValues returns the allowed values for the given BR-CL rule, or nil if
// none is known.
func CodeListValues(ruleID string) []string {
	list, ok := codeLists[canon(ruleID)]
	if !ok {
		return nil
	}
	return strings.Fields(list)
}

// CodeLists returns the number of code lists in the catalog.
func CodeLists() int { return len(codeLists) }
