// Package xinvoice converts German XRechnung 3.0.2 electronic invoices between
// a syntax-neutral JSON model and the two official XML syntaxes (UBL and
// UN/CEFACT CII), and validates them against the EN16931 and XRechnung business
// rules.
//
// The public surface is intentionally small:
//
//   - [ParseXML] reads UBL or CII XML (auto-detected) into an [Invoice].
//   - [Invoice.ToXML] serializes an [Invoice] back to a chosen [Syntax].
//   - [FromJSON] and [Invoice.ToJSON] move between JSON and the model.
//   - [Validate] and [ValidateXML] return a [ValidationResult] whose findings
//     can be rendered as JSON in German or English via [ValidationResult.JSON].
//
// The [Invoice] model mirrors the EN16931 semantic model (Business Terms and
// Groups), so a single JSON document maps to either XML syntax.
package xinvoice
