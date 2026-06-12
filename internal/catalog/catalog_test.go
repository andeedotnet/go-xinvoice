package catalog

import "testing"

func TestCatalogPopulated(t *testing.T) {
	if Rules() < 200 {
		t.Errorf("expected a few hundred rules, got %d", Rules())
	}
	if Labels() < 150 {
		t.Errorf("expected ~200 labels, got %d", Labels())
	}
}

func TestRuleMergeAcrossSources(t *testing.T) {
	// BR-1 exists in the SeMoX model (German) and the EN16931 XSL (English, as
	// "BR-01"); the id-canonicalization must merge them into one rule that has
	// both languages and a severity.
	r, ok := LookupRule("BR-1")
	if !ok {
		t.Fatal("BR-1 not found")
	}
	if r.DE == "" || r.EN == "" {
		t.Errorf("BR-1 not merged across sources: DE=%q EN=%q", r.DE, r.EN)
	}
	if r.Severity != SeverityError {
		t.Errorf("BR-1 severity = %q, want error", r.Severity)
	}
	// Lookups accept the padded EN16931 form too and resolve to the same rule.
	if padded, ok := LookupRule("BR-01"); !ok || padded.ID != r.ID {
		t.Errorf("LookupRule(BR-01) did not resolve to BR-1: %+v ok=%v", padded, ok)
	}
}

func TestRuleTextFallback(t *testing.T) {
	// BR-CL-1 (a code-list rule) is English-only; the German request must fall
	// back to the English text rather than returning empty.
	r, ok := LookupRule("BR-CL-1")
	if !ok {
		t.Fatal("BR-CL-1 not found")
	}
	if r.EN == "" {
		t.Fatal("BR-CL-1 expected to have English text")
	}
	if got := r.Text("de"); got != r.EN {
		t.Errorf("Text(de) fallback = %q, want English %q", got, r.EN)
	}
}

func TestBRDEIsGermanWithSeverity(t *testing.T) {
	r, ok := LookupRule("BR-DE-1")
	if !ok {
		t.Fatal("BR-DE-1 not found")
	}
	if r.DE == "" {
		t.Error("BR-DE-1 expected German text")
	}
	if r.Severity == "" {
		t.Error("BR-DE-1 expected a severity")
	}
}

func TestLabelLookup(t *testing.T) {
	// BT-1 = Invoice number / Rechnungsnummer.
	l, ok := LookupLabel("BT-1")
	if !ok {
		t.Fatal("BT-1 label not found")
	}
	if l.Text("de") != "Rechnungsnummer" {
		t.Errorf("BT-1 de label = %q", l.Text("de"))
	}
	if l.Text("en") == "" {
		t.Errorf("BT-1 en label empty")
	}
}

func TestCodeListsPopulated(t *testing.T) {
	if CodeLists() < 20 {
		t.Errorf("expected ~22 code lists, got %d", CodeLists())
	}
}

func TestCodeListMembership(t *testing.T) {
	cases := []struct {
		rule, value string
		want        bool
	}{
		{"BR-CL-17", "S", true},    // VAT category (UNCL5305)
		{"BR-CL-17", "Z", true},    //
		{"BR-CL-17", "XX", false},  //
		{"BR-CL-04", "EUR", true},  // invoice currency (ISO 4217)
		{"BR-CL-04", "USD", true},  //
		{"BR-CL-04", "XXX", true},  // ISO 4217 "no currency" — legitimately present
		{"BR-CL-04", "ZZZ", false}, // not an ISO 4217 code
		{"BR-CL-14", "DE", true},   // country (ISO 3166)
		{"BR-CL-14", "ZZ", false},
		{"BR-CL-01", "380", true}, // document type (UNTDID 1001)
		{"BR-CL-01", "381", true}, //
		{"BR-CL-01", "999", false},
		{"BR-CL-23", "C62", true}, // unit code (Rec 20)
		{"BR-CL-16", "58", true},  // payment means (UNCL4461)
	}
	for _, c := range cases {
		got, known := InCodeList(c.rule, c.value)
		if !known {
			t.Errorf("%s: list unknown", c.rule)
			continue
		}
		if got != c.want {
			t.Errorf("InCodeList(%s, %q) = %v, want %v", c.rule, c.value, got, c.want)
		}
	}

	// A value with surrounding spaces (as the schematron normalizes) still matches.
	if in, _ := InCodeList("BR-CL-17", "  S  "); !in {
		t.Error("expected trimmed membership match")
	}
	// BR-CL-24 has no enumerated list.
	if _, known := InCodeList("BR-CL-24", "application/pdf"); known {
		t.Error("BR-CL-24 should not have an enumerated list")
	}
}

func TestCodeListSizes(t *testing.T) {
	// Sanity-check a couple of well-known list sizes against the EN16931 source.
	if n := len(CodeListValues("BR-CL-04")); n < 170 || n > 190 {
		t.Errorf("ISO 4217 currency list size = %d, want ~179", n)
	}
	if n := len(CodeListValues("BR-CL-14")); n < 240 || n > 260 {
		t.Errorf("ISO 3166 country list size = %d, want ~251", n)
	}
	if vat := CodeListValues("BR-CL-17"); len(vat) < 8 {
		t.Errorf("VAT category list too small: %v", vat)
	}
}

func TestRuleTermsLinkToLabels(t *testing.T) {
	// BR-2 ("invoice must have a number") is about BT-1, which must resolve to a
	// label — proving rule.Terms and the label catalog share the same id space.
	r, ok := LookupRule("BR-2")
	if !ok {
		t.Fatal("BR-2 not found")
	}
	if len(r.Terms) == 0 {
		t.Fatal("BR-2 expected at least one term")
	}
	for _, term := range r.Terms {
		if _, ok := LookupLabel(term); !ok {
			t.Errorf("BR-2 term %q has no label", term)
		}
	}
}
