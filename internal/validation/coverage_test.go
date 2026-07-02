package validation

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/andeedotnet/go-xinvoice/internal/catalog"
)

// outOfScope lists catalog rules that are intentionally not implemented, by id
// prefix, each with a reason. These are rules that cannot be meaningfully checked
// against the syntax-neutral model.
var outOfScope = []struct{ prefix, reason string }{
	{"BR-DEX-", "extension sub-invoice-line rules (deep extension validation)"},
	{"BR-AF-", "non-German VAT category (no testsuite data to verify against)"},
	{"BR-AG-", "non-German VAT category (no testsuite data to verify against)"},
	{"BR-B-", "split payment is an Italian domestic category"},
	{"BR-DE-CVD-", "XRechnung CVD construction profile"},
	{"BR-TMP-CVD-", "CVD construction profile (temporary rule)"},
	{"BR-DE-TMP-", "temporary rule"},
	{"BR-TMP-", "temporary rule"},
	{"BR-CO-5", "allowance reason code↔reason type equivalence (no mapping table)"},
	{"BR-CO-6", "charge reason code↔reason type equivalence (no mapping table)"},
	{"BR-CO-7", "line allowance reason code↔reason type equivalence"},
	{"BR-CO-8", "line charge reason code↔reason type equivalence"},
	{"BR-DE-12", "SeMoX model only (no schematron assertion); BT-78 post-code presence is already covered by BR-DE-11 and cannot be verified independently"},
	{"BR-DE-14", "covered structurally by BR-48 (VAT rate present)"},
	{"BR-DE-18", "skonto free-text format (regex over BT-20)"},
	{"BR-DE-23", "payment-means-conditional account-detail rules"},
	{"BR-DE-24", "payment-means-conditional account-detail rules"},
	{"BR-DE-25", "payment-means-conditional account-detail rules"},
	{"BR-DE-29", "deprecated, replaced by PEPPOL-EN16931-R061"},
	{"BR-51", "full card PAN masking is not detectable from the model"},
	{"BR-CL-24", "MIME type is validated by regex, not an enumerated list"},
	{"PEPPOL-EN16931-R008", "syntax-level: document must not contain empty elements"},
	{"PEPPOL-EN16931-R043", "syntax-level: charge indicator value"},
	{"PEPPOL-EN16931-R044", "structurally satisfied: a price-level charge cannot be modelled"},
	{"PEPPOL-EN16931-R053", "syntax-level: one tax total with subtotals"},
	{"PEPPOL-EN16931-R054", "syntax-level: one tax total without subtotals"},
	{"PEPPOL-EN16931-R101", "syntax-level: DocumentReference usage"},
}

func outOfScopeReason(id string) (string, bool) {
	for _, o := range outOfScope {
		if strings.HasPrefix(id, o.prefix) {
			return o.reason, true
		}
	}
	return "", false
}

// TestRuleCoverage fails if a catalog rule is neither implemented (emitted by a
// rules_*.go file) nor explicitly listed as out of scope. It keeps coverage
// honest and flags new rules after a bundle update.
func TestRuleCoverage(t *testing.T) {
	impl := emittedRuleIDs(t)

	var uncovered []string
	covered := 0
	for _, id := range catalog.RuleIDs() {
		switch {
		case impl[id]:
			covered++
		case func() bool { _, ok := outOfScopeReason(id); return ok }():
			// documented gap
		default:
			uncovered = append(uncovered, id)
		}
	}
	total := len(catalog.RuleIDs())
	t.Logf("rule coverage: %d/%d implemented (%d%%), %d out-of-scope, %d uncovered",
		covered, total, covered*100/total, total-covered-len(uncovered), len(uncovered))
	if len(uncovered) > 0 {
		t.Errorf("uncovered rules (implement them or add to outOfScope): %s", strings.Join(uncovered, " "))
	}
}

// emittedRuleIDs derives the set of rule ids the validator can emit by scanning
// the rules_*.go sources for rule-id literals plus the VAT-category family
// expansion (those ids are built dynamically as family+"-N").
func emittedRuleIDs(t *testing.T) map[string]bool {
	out := map[string]bool{}
	idRe := regexp.MustCompile(`"((?:BR|PEPPOL)[A-Za-z0-9-]+)"`)
	files, _ := filepath.Glob("rules_*.go")
	if len(files) == 0 {
		t.Skip("rule sources not found")
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, m := range idRe.FindAllStringSubmatch(string(data), -1) {
			out[canon(m[1])] = true
		}
	}
	// VAT category families (rules_vat.go builds ids as family+"-N").
	for _, fam := range []string{"BR-S", "BR-Z", "BR-E", "BR-AE", "BR-IC", "BR-G", "BR-O", "BR-IG", "BR-IP"} {
		for n := 1; n <= 10; n++ {
			out[fam+"-"+strconv.Itoa(n)] = true
		}
	}
	for _, id := range []string{"BR-IC-11", "BR-IC-12", "BR-O-11", "BR-O-12", "BR-O-13", "BR-O-14"} {
		out[id] = true
	}
	return out
}

// canon mirrors the catalog id canonicalization (strip leading zeros in numeric
// segments) so "BR-CL-04" and "BR-CL-4" match.
func canon(id string) string {
	parts := strings.Split(id, "-")
	for i, p := range parts {
		if n, err := strconv.Atoi(p); err == nil && p != "" {
			parts[i] = strconv.Itoa(n)
		}
	}
	return strings.Join(parts, "-")
}
