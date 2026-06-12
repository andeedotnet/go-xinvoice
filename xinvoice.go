package xinvoice

import (
	"fmt"

	"github.com/andeedotnet/go-xinvoice/internal/catalog"
	"github.com/andeedotnet/go-xinvoice/internal/checksum"
	"github.com/andeedotnet/go-xinvoice/internal/cii"
	"github.com/andeedotnet/go-xinvoice/internal/ubl"
	"github.com/andeedotnet/go-xinvoice/internal/validation"
	"github.com/andeedotnet/go-xinvoice/model"
)

// ValidationResult holds the findings from validating an invoice. Render them as
// JSON in German or English with [ValidationResult.JSON], or inspect the typed
// [ValidationResult.Findings] directly.
type ValidationResult = validation.Result

// Finding is a single rule violation (see [ValidationResult.Findings]).
type Finding = validation.Finding

// Severity classifies how a violated rule is reported.
type Severity = catalog.Severity

// Severity levels.
const (
	SeverityError   = catalog.SeverityError
	SeverityWarning = catalog.SeverityWarning
	SeverityInfo    = catalog.SeverityInfo
)

// Check-digit validators for identifiers used in invoices (GS1 GLN mod-10, IBAN
// ISO-13616 mod-97, generic mod-11).
var (
	ValidGLN  = checksum.ValidGLN
	ValidIBAN = checksum.ValidIBAN
	Mod11     = checksum.Mod11
)

// Validate checks an [Invoice] against the EN16931 / XRechnung business rules.
func Validate(inv *Invoice) *ValidationResult { return validation.Validate(inv) }

// ValidateXML parses UBL or CII XML (auto-detected) and validates it. A parse
// failure is reported as a single finding rather than a separate error.
func ValidateXML(b []byte) *ValidationResult {
	inv, err := ParseXML(b)
	if err != nil {
		return validation.ParseErrorResult(err)
	}
	return validation.Validate(inv)
}

func init() {
	model.RegisterXMLCodec(toXML, parseXML)
}

func toXML(inv *model.Invoice, s model.Syntax) ([]byte, error) {
	switch s {
	case model.UBL:
		return ubl.Marshal(inv)
	case model.CII:
		return cii.Marshal(inv)
	default:
		return nil, fmt.Errorf("xinvoice: unknown syntax %v", s)
	}
}

func parseXML(b []byte) (*model.Invoice, error) {
	switch {
	case ubl.Detect(b):
		return ubl.Parse(b)
	case cii.Detect(b):
		return cii.Parse(b)
	default:
		return nil, fmt.Errorf("xinvoice: %w: unrecognized XML (neither UBL nor CII)", model.ErrNotImplemented)
	}
}

// Invoice is the syntax-neutral XRechnung invoice (the EN16931 semantic model).
// See the [model] package for the full field-by-field BT/BG documentation.
type Invoice = model.Invoice

// Syntax selects one of the two official XRechnung XML representations.
type Syntax = model.Syntax

const (
	// UBL is the OASIS UBL 2.1 syntax.
	UBL = model.UBL
	// CII is the UN/CEFACT Cross Industry Invoice syntax.
	CII = model.CII
)

// ErrNotImplemented is returned by surface area declared ahead of its
// implementation (the XML codecs and validator land in later phases).
var ErrNotImplemented = model.ErrNotImplemented

// FromJSON parses a semantic-model JSON document into an [Invoice].
func FromJSON(b []byte) (*Invoice, error) { return model.FromJSON(b) }

// ParseXML reads UBL or CII XML (auto-detected) into an [Invoice].
func ParseXML(b []byte) (*Invoice, error) { return model.ParseXML(b) }
