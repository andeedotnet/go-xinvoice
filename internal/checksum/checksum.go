// Package checksum provides check-digit validators for identifiers used in
// XRechnung invoices: GS1 GLN (mod-10), IBAN (ISO 13616 mod-97) and a generic
// ISO-7064-style mod-11.
//
// These validate the syntactic check digits only, not whether the identifier
// actually exists. They are exposed as utilities; the validator does not enforce
// them by default, because the official testsuite uses placeholder identifiers
// that would fail a strict check.
package checksum

import "strings"

// ValidGLN reports whether s is a syntactically valid 13-digit GS1 Global
// Location Number (scheme 0088), i.e. its mod-10 check digit is correct. This
// mirrors the official schematron's u:gln function.
func ValidGLN(s string) bool {
	if len(s) != 13 || !allDigits(s) {
		return false
	}
	sum := 0
	for k := 0; k < 12; k++ {
		d := int(s[11-k] - '0') // iterate the 12-digit body from the right
		weight := 1
		if k%2 == 0 {
			weight = 3
		}
		sum += d * weight
	}
	check := (10 - sum%10) % 10
	return check == int(s[12]-'0')
}

// ValidIBAN reports whether s is a valid IBAN per the ISO 13616 mod-97 check
// (spaces are ignored). It validates the check digits, not whether the account
// exists.
func ValidIBAN(s string) bool {
	s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	if len(s) < 5 || len(s) > 34 {
		return false
	}
	// Letters in the first two positions (country), digits in positions 3-4.
	if !isAlpha(s[0]) || !isAlpha(s[1]) || !isDigit(s[2]) || !isDigit(s[3]) {
		return false
	}
	rearranged := s[4:] + s[:4]
	rem := 0
	for i := 0; i < len(rearranged); i++ {
		c := rearranged[i]
		switch {
		case isDigit(c):
			rem = (rem*10 + int(c-'0')) % 97
		case isAlpha(c):
			// A=10 .. Z=35 — two digits.
			v := int(c-'A') + 10
			rem = (rem*100 + v) % 97
		default:
			return false
		}
	}
	return rem == 1
}

// Mod11 reports whether the last digit of s is a valid ISO-7064-style mod-11
// check digit over the preceding digits (used by some registration ids).
func Mod11(s string) bool {
	if len(s) < 2 || !allDigits(s) {
		return false
	}
	body, check := s[:len(s)-1], int(s[len(s)-1]-'0')
	sum, weight := 0, 2
	for i := len(body) - 1; i >= 0; i-- {
		sum += int(body[i]-'0') * weight
		weight++
		if weight > 7 {
			weight = 2
		}
	}
	r := (11 - sum%11) % 11
	if r == 10 {
		return false // not representable as a single digit
	}
	return r == check
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) {
			return false
		}
	}
	return len(s) > 0
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }
func isAlpha(b byte) bool { return b >= 'A' && b <= 'Z' }
