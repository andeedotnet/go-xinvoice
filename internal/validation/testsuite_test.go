package validation_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/andeedotnet/go-xinvoice/internal/cii"
	"github.com/andeedotnet/go-xinvoice/internal/ubl"
	"github.com/andeedotnet/go-xinvoice/internal/validation"
	"github.com/andeedotnet/go-xinvoice/model"
)

// TestTestsuiteClean validates every standard / technical-case instance (both
// UBL and CII) and requires zero error-severity findings: the official testsuite
// is valid, so any error is a false positive in our rules. Warnings are allowed.
func TestTestsuiteClean(t *testing.T) {
	root := findTestsuite(t)

	type parsed struct {
		name string
		inv  *model.Invoice
	}
	var docs []parsed
	for _, sub := range []string{"standard", "technical-cases"} {
		_ = filepath.WalkDir(filepath.Join(root, sub), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			raw, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			switch {
			case strings.HasSuffix(path, "_ubl.xml"):
				if strings.Contains(string(raw), "SubInvoiceLine") {
					return nil // extension, not modeled
				}
				if inv, e := ubl.Parse(raw); e == nil {
					docs = append(docs, parsed{filepath.Base(path), inv})
				}
			case strings.HasSuffix(path, "_uncefact.xml"):
				if inv, e := cii.Parse(raw); e == nil {
					docs = append(docs, parsed{filepath.Base(path), inv})
				}
			}
			return nil
		})
	}
	if len(docs) == 0 {
		skipNoTestsuite(t, "no testsuite instances found")
	}

	falsePositives := map[string]int{} // rule -> count
	failedDocs := 0
	for _, d := range docs {
		res := validation.Validate(d.inv)
		var errs []string
		for _, f := range res.Findings {
			if f.Severity == "error" {
				falsePositives[f.Rule]++
				errs = append(errs, f.Rule+"("+f.Location+" "+f.Detail+")")
			}
		}
		if len(errs) > 0 {
			failedDocs++
			if failedDocs <= 8 {
				t.Errorf("%s: unexpected errors: %s", d.name, strings.Join(errs, ", "))
			}
		}
	}
	if len(falsePositives) > 0 {
		var rules []string
		for r := range falsePositives {
			rules = append(rules, r)
		}
		sort.Strings(rules)
		var summary []string
		for _, r := range rules {
			summary = append(summary, r+"="+itoa(falsePositives[r]))
		}
		t.Errorf("false positives across %d/%d docs by rule: %s", failedDocs, len(docs), strings.Join(summary, " "))
	} else {
		t.Logf("validated %d instances clean (no error findings)", len(docs))
	}
}

func findTestsuite(t *testing.T) string {
	dir, _ := os.Getwd()
	for {
		if m, _ := filepath.Glob(filepath.Join(dir, "knowledge", "*", "*-testsuite-*", "instances")); len(m) > 0 {
			return m[0]
		}
		if td := filepath.Join(dir, "testdata", "instances"); isDir(td) {
			return td
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	skipNoTestsuite(t, "no testdata/ or knowledge/ instances present")
	return ""
}

// skipNoTestsuite skips when no testsuite instances are present — unless
// XINVOICE_REQUIRE_TESTSUITE is set (CI sets it), in which case their absence is
// a hard failure so the zero-false-positives verification can never pass merely
// by being skipped.
func skipNoTestsuite(t *testing.T, msg string) {
	t.Helper()
	if os.Getenv("XINVOICE_REQUIRE_TESTSUITE") != "" {
		t.Fatalf("XINVOICE_REQUIRE_TESTSUITE is set but %s", msg)
	}
	t.Skip(msg)
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
