package validation

import (
	"math/big"
	"strings"

	"github.com/andeedotnet/go-xinvoice/model"
)

// present reports whether s carries a non-whitespace value.
func present(s string) bool { return strings.TrimSpace(s) != "" }

// rat returns the exact value of d, or 0 for an empty/invalid decimal.
func rat(d model.Decimal) *big.Rat {
	r, ok := d.Rat()
	if !ok || r == nil {
		return new(big.Rat)
	}
	return r
}

// zero is a reusable big.Rat zero.
func zero() *big.Rat { return new(big.Rat) }

func add(a, b *big.Rat) *big.Rat { return new(big.Rat).Add(a, b) }
func sub(a, b *big.Rat) *big.Rat { return new(big.Rat).Sub(a, b) }
func mul(a, b *big.Rat) *big.Rat { return new(big.Rat).Mul(a, b) }

// slack returns the rounding tolerance for an invoice currency: 0.5 for HUF,
// else 0.02 — matching the official schematron's slack value.
func slack(currency string) *big.Rat {
	if currency == "HUF" {
		return big.NewRat(1, 2)
	}
	return big.NewRat(2, 100)
}

// within reports whether actual is within tol of expected.
func within(expected, actual, tol *big.Rat) bool {
	diff := new(big.Rat).Sub(expected, actual)
	diff.Abs(diff)
	return diff.Cmp(tol) <= 0
}

// f2 formats a rat as a 2-decimal string for finding details.
func f2(r *big.Rat) string { return r.FloatString(2) }
