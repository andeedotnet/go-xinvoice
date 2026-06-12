package cii

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestTestsuiteRoundTrip parses every standard / technical-case CII instance
// from the official KoSIT testsuite, serializes it back to CII and re-parses it,
// requiring the two models to be identical. Instances come from the curated
// testdata/, or the full upstream bundle dropped into knowledge/ (downloaded from
// https://xeinkauf.de/xrechnung/versionen-und-bundles/); it skips when neither is
// present.
func TestTestsuiteRoundTrip(t *testing.T) {
	dir := findTestsuite(t)
	files := findCIIInstances(dir)
	if len(files) == 0 {
		skipNoTestsuite(t, "no CII testsuite instances found")
	}

	tested := 0
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		tested++
		t.Run(filepath.Base(f), func(t *testing.T) {
			m1, err := Parse(raw)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			again, err := Marshal(m1)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			m2, err := Parse(again)
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			if !reflect.DeepEqual(m1, m2) {
				j1, _ := m1.ToJSON()
				j2, _ := m2.ToJSON()
				t.Errorf("round-trip not idempotent.\nfirst diff: %s", firstDiff(string(j1), string(j2)))
			}
		})
	}
	t.Logf("round-tripped %d CII instances", tested)
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
// a hard failure so the round-trip verification can never pass merely by being
// skipped.
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

func findCIIInstances(root string) []string {
	var out []string
	for _, sub := range []string{"standard", "technical-cases"} {
		_ = filepath.WalkDir(filepath.Join(root, sub), func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, "_uncefact.xml") {
				out = append(out, path)
			}
			return nil
		})
	}
	return out
}

func firstDiff(a, b string) string {
	la, lb := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := 0; i < len(la) || i < len(lb); i++ {
		var x, y string
		if i < len(la) {
			x = la[i]
		}
		if i < len(lb) {
			y = lb[i]
		}
		if x != y {
			return "line " + itoa(i+1) + "\n  parsed:   " + strings.TrimSpace(x) + "\n  reparsed: " + strings.TrimSpace(y)
		}
	}
	return "(no line diff; structural/ordering difference)"
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
