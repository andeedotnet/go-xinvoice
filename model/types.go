// Package model defines the syntax-neutral XRechnung 3.0.2 semantic model.
//
// The types here mirror the EN16931 semantic model: Business Terms (BT-*) and
// Business Groups (BG-*). A single [Invoice] value therefore maps to either of
// the two official XML syntaxes (UBL or UN/CEFACT CII) and to a stable JSON
// document. Every field on [Invoice] and its sub-structs is annotated with the
// BT/BG identifier it represents.
package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"
)

// Syntax selects one of the two official XRechnung XML representations.
type Syntax int

const (
	// UBL is the OASIS UBL 2.1 Invoice / CreditNote syntax.
	UBL Syntax = iota
	// CII is the UN/CEFACT Cross Industry Invoice (CII) D16B syntax.
	CII
)

func (s Syntax) String() string {
	switch s {
	case UBL:
		return "UBL"
	case CII:
		return "CII"
	default:
		return fmt.Sprintf("Syntax(%d)", int(s))
	}
}

// ErrNotImplemented is returned by surface area that is declared but not yet
// wired up (the XML codecs and the validator are added in later phases).
var ErrNotImplemented = errors.New("xinvoice: not implemented")

// XML codec hooks. The internal/ubl, internal/cii and internal/validation
// packages register their implementations here during init, which keeps this
// model package a dependency-free leaf (no import cycle).
var (
	toXMLFunc    func(*Invoice, Syntax) ([]byte, error)
	parseXMLFunc func([]byte) (*Invoice, error)
)

// RegisterXMLCodec wires the XML serializer and parser. It is called from
// package init in the syntax packages and is not meant for direct use.
func RegisterXMLCodec(to func(*Invoice, Syntax) ([]byte, error), parse func([]byte) (*Invoice, error)) {
	toXMLFunc, parseXMLFunc = to, parse
}

// ToXML serializes the invoice to the requested official XML [Syntax].
func (inv *Invoice) ToXML(s Syntax) ([]byte, error) {
	if toXMLFunc == nil {
		return nil, ErrNotImplemented
	}
	return toXMLFunc(inv, s)
}

// ParseXML reads UBL or CII XML (auto-detected) into an [Invoice].
func ParseXML(b []byte) (*Invoice, error) {
	if parseXMLFunc == nil {
		return nil, ErrNotImplemented
	}
	return parseXMLFunc(b)
}

// ---------------------------------------------------------------------------
// Decimal-based scalar types
// ---------------------------------------------------------------------------

var decimalRe = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

// Decimal is an exact decimal number held in its lexical form (so trailing
// zeros such as "1000.00" are preserved). It marshals to JSON as a number and
// unmarshals from either a JSON number or a JSON string. The empty Decimal
// represents "not set" and is dropped by `omitempty`.
type Decimal string

// Amount is a monetary [Decimal]. Unless stated otherwise it is expressed in
// the invoice currency (BT-5); the only exceptions are BT-6/BT-111 (accounting
// currency). The currency is therefore not carried per amount in the model.
type Amount = Decimal

// Percentage is a [Decimal] expressing a percentage value (e.g. a VAT rate).
type Percentage = Decimal

// ParseDecimal validates s and returns it as a [Decimal].
func ParseDecimal(s string) (Decimal, error) {
	if !decimalRe.MatchString(s) {
		return "", fmt.Errorf("xinvoice: %q is not a valid decimal", s)
	}
	return Decimal(s), nil
}

// Rat returns the exact value as a *big.Rat. The bool reports whether the
// Decimal held a parseable value (false for the empty Decimal).
func (d Decimal) Rat() (*big.Rat, bool) {
	if d == "" {
		return nil, false
	}
	r := new(big.Rat)
	_, ok := r.SetString(string(d))
	return r, ok
}

// String returns the lexical form.
func (d Decimal) String() string { return string(d) }

// MarshalJSON emits the decimal as a bare JSON number.
func (d Decimal) MarshalJSON() ([]byte, error) {
	if d == "" {
		return []byte("null"), nil
	}
	if !decimalRe.MatchString(string(d)) {
		return nil, fmt.Errorf("xinvoice: %q is not a valid decimal", string(d))
	}
	return []byte(d), nil
}

// UnmarshalJSON accepts a JSON number or a JSON string containing a decimal.
func (d *Decimal) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if string(b) == "null" {
		*d = ""
		return nil
	}
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	if !decimalRe.MatchString(s) {
		return fmt.Errorf("xinvoice: %q is not a valid decimal", s)
	}
	*d = Decimal(s)
	return nil
}

// ---------------------------------------------------------------------------
// Date
// ---------------------------------------------------------------------------

// Date is an EN16931 calendar date in ISO "YYYY-MM-DD" form (no time, no zone).
// It marshals to/from a JSON string. The empty Date means "not set".
type Date string

// NewDate formats a time.Time as a [Date] (its date component, UTC-agnostic).
func NewDate(t time.Time) Date { return Date(t.Format("2006-01-02")) }

// Time parses the Date into a time.Time at midnight UTC.
func (d Date) Time() (time.Time, error) { return time.Parse("2006-01-02", string(d)) }

// MarshalJSON validates the format and emits a JSON string.
func (d Date) MarshalJSON() ([]byte, error) {
	if d == "" {
		return []byte("null"), nil
	}
	if _, err := d.Time(); err != nil {
		return nil, fmt.Errorf("xinvoice: %q is not a valid YYYY-MM-DD date", string(d))
	}
	return json.Marshal(string(d))
}

// UnmarshalJSON reads a JSON string and validates the date format.
func (d *Date) UnmarshalJSON(b []byte) error {
	if string(bytes.TrimSpace(b)) == "null" {
		*d = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != "" {
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return fmt.Errorf("xinvoice: %q is not a valid YYYY-MM-DD date", s)
		}
	}
	*d = Date(s)
	return nil
}

// ---------------------------------------------------------------------------
// Code
// ---------------------------------------------------------------------------

// Code is a coded value taken from a code list. The applicable list is fixed by
// the Business Term that carries the code (e.g. BT-3 invoice type, BT-40 country
// code), so the list identifier is not stored alongside the value.
type Code = string

// ---------------------------------------------------------------------------
// Quantity
// ---------------------------------------------------------------------------

// Quantity is a measured amount together with its unit of measure code
// (UN/ECE Recommendation 20 / 21). It folds a value BT and its sibling unit BT
// into one field, e.g. BT-129 with BT-130, or BT-149 with BT-150.
type Quantity struct {
	Value Decimal `json:"value"`
	Unit  Code    `json:"unit,omitempty"`
}

// ---------------------------------------------------------------------------
// ID (identifier with optional scheme)
// ---------------------------------------------------------------------------

// ID is an identifier with an optional scheme. Many Business Terms pair a value
// with a "-1" scheme sub-term (e.g. BT-29 with BT-29-1). When Scheme is empty
// the ID marshals to a plain JSON string; otherwise it marshals to an object
// {"value":..., "scheme":...}. Both forms are accepted on input.
type ID struct {
	Value  string `json:"value"`
	Scheme Code   `json:"scheme,omitempty"`
}

// MarshalJSON emits a bare string when there is no scheme, else an object.
func (id ID) MarshalJSON() ([]byte, error) {
	if id.Scheme == "" {
		return json.Marshal(id.Value)
	}
	type alias ID
	return json.Marshal(alias(id))
}

// UnmarshalJSON accepts either a bare string or an object form.
func (id *ID) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		id.Value, id.Scheme = s, ""
		return nil
	}
	type alias ID
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*id = ID(a)
	return nil
}

// ---------------------------------------------------------------------------
// BinaryObject (attached document)
// ---------------------------------------------------------------------------

// BinaryObject is an embedded binary file, used by BT-125 (attached document)
// together with its MIME code (BT-125-1) and filename (BT-125-2). Content is
// base64-encoded in JSON.
type BinaryObject struct {
	MimeCode Code   `json:"mimeCode,omitempty"`
	Filename string `json:"filename,omitempty"`
	Content  []byte `json:"content,omitempty"`
}
