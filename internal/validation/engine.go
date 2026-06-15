// Package validation reimplements the EN16931 / XRechnung business rules
// natively in Go against the syntax-neutral [model.Invoice]. Because both UBL
// and CII map to the same model, one rule set covers both syntaxes.
//
// Rule texts, severities and code lists come from the generated
// internal/catalog (derived from the official KoSIT sources). Coverage grows
// incrementally; see TODO.md for the per-category status.
package validation

import (
	"encoding/json"
	"sort"

	"github.com/andeedotnet/go-xinvoice/internal/catalog"
	"github.com/andeedotnet/go-xinvoice/model"
)

// Finding is a single rule violation.
type Finding struct {
	Rule     string
	Severity catalog.Severity
	Location string // BT/BG identifier the finding is about (optional)
	Detail   string // rule-specific specifics, e.g. computed vs expected (optional)
}

// Result is the outcome of validating an invoice.
type Result struct {
	Findings []Finding
}

// Valid reports whether the invoice has no error-severity findings (warnings and
// informational findings do not make it invalid).
func (r *Result) Valid() bool {
	for _, f := range r.Findings {
		if f.Severity == catalog.SeverityError {
			return false
		}
	}
	return true
}

// JSON renders the result with rule messages in the requested language
// ("de" or "en").
func (r *Result) JSON(lang string) ([]byte, error) {
	type finding struct {
		Rule          string `json:"rule"`
		Severity      string `json:"severity"`
		Location      string `json:"location,omitempty"`
		LocationLabel string `json:"locationLabel,omitempty"`
		Message       string `json:"message"`
		Detail        string `json:"detail,omitempty"`
	}
	out := struct {
		Valid    bool      `json:"valid"`
		Findings []finding `json:"findings"`
	}{Valid: r.Valid(), Findings: []finding{}}

	for _, f := range r.Findings {
		msg := f.Rule
		if ru, ok := catalog.LookupRule(f.Rule); ok {
			if t := ru.Text(lang); t != "" {
				msg = t
			}
		}
		// Resolve the BT/BG term label (e.g. BT-10 -> "Käuferreferenz") so callers
		// can show what the location field is.
		label := ""
		if f.Location != "" {
			if l, ok := catalog.LookupLabel(f.Location); ok {
				label = l.Text(lang)
			}
		}
		out.Findings = append(out.Findings, finding{
			Rule: f.Rule, Severity: string(f.Severity), Location: f.Location, LocationLabel: label, Message: msg, Detail: f.Detail,
		})
	}
	return json.MarshalIndent(out, "", "  ")
}

// ---------------------------------------------------------------------------
// Reporter + registry
// ---------------------------------------------------------------------------

type reporter struct {
	findings []Finding
}

// fail records a violation of rule, located at the given BT/BG.
func (rp *reporter) fail(rule, location string) {
	rp.failf(rule, location, "")
}

// failf records a violation with extra detail.
func (rp *reporter) failf(rule, location, detail string) {
	sev := catalog.SeverityError
	if r, ok := catalog.LookupRule(rule); ok && r.Severity != "" {
		sev = r.Severity
	}
	rp.findings = append(rp.findings, Finding{Rule: rule, Severity: sev, Location: location, Detail: detail})
}

type ruleFunc func(*model.Invoice, *reporter)

var registry []ruleFunc

// register adds rule functions to the global registry (called from init in the
// rules_*.go files).
func register(fns ...ruleFunc) { registry = append(registry, fns...) }

// ParseErrorResult builds a result carrying a single XML parse error, so that
// the facade's ValidateXML can report unparseable input without a separate error
// return.
func ParseErrorResult(err error) *Result {
	return &Result{Findings: []Finding{{Rule: "XML-PARSE", Severity: catalog.SeverityError, Detail: err.Error()}}}
}

// Validate runs every registered rule against inv and returns the findings,
// sorted by rule id then location for deterministic output.
func Validate(inv *model.Invoice) *Result {
	rp := &reporter{}
	for _, fn := range registry {
		fn(inv, rp)
	}
	sort.SliceStable(rp.findings, func(i, j int) bool {
		if rp.findings[i].Rule != rp.findings[j].Rule {
			return ruleLess(rp.findings[i].Rule, rp.findings[j].Rule)
		}
		return rp.findings[i].Location < rp.findings[j].Location
	})
	return &Result{Findings: rp.findings}
}

// ruleLess orders rule ids by their family then numeric suffix (so BR-2 sorts
// before BR-10).
func ruleLess(a, b string) bool {
	fa, na := splitRule(a)
	fb, nb := splitRule(b)
	if fa != fb {
		return fa < fb
	}
	if na != nb {
		return na < nb
	}
	return a < b
}

func splitRule(id string) (family string, num int) {
	// Trailing run of digits is the numeric part; the rest is the family.
	i := len(id)
	for i > 0 && id[i-1] >= '0' && id[i-1] <= '9' {
		i--
	}
	family = id[:i]
	for _, c := range id[i:] {
		num = num*10 + int(c-'0')
	}
	return family, num
}
