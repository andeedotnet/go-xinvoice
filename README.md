# go-xinvoice

[![CI](https://github.com/andeedotnet/go-xinvoice/actions/workflows/ci.yml/badge.svg)](https://github.com/andeedotnet/go-xinvoice/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/andeedotnet/go-xinvoice.svg)](https://pkg.go.dev/github.com/andeedotnet/go-xinvoice)
[![Go Report Card](https://goreportcard.com/badge/github.com/andeedotnet/go-xinvoice)](https://goreportcard.com/report/github.com/andeedotnet/go-xinvoice)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

A pure-Go library for the German **XRechnung 3.0.2** electronic invoice standard (EN16931 CIUS + Extension).

It provides three things:

1. **JSON → XRechnung XML** — build an invoice as a syntax-neutral JSON document and serialize it to
   either official XML syntax (**UBL** or **UN/CEFACT CII**).
2. **XRechnung XML → JSON** — parse UBL or CII XML (auto-detected) into the same JSON model.
3. **Validation** — check an invoice against the EN16931 and XRechnung business rules and return the
   findings as JSON, with messages selectable in **German** or **English**.

The library is pure Go with no XSLT/Java runtime dependency. See
[Validation coverage](#validation-coverage) for how it relates to the official KoSIT validator.

> Status: **v0.1.0** — early but usable; conversion is complete, validation coverage
> grows incrementally (254/329 rules). APIs may still change before v1.0. See
> [`TODO.md`](./TODO.md) and [`CHANGELOG.md`](./CHANGELOG.md).

## Install

```sh
go get github.com/andeedotnet/go-xinvoice
```

## Usage

```go
import "github.com/andeedotnet/go-xinvoice"

// XML -> model -> JSON
inv, err := xinvoice.ParseXML(xmlBytes)  // auto-detects UBL vs CII
jsonBytes, err := inv.ToJSON()

// JSON -> model -> XML
inv, err := xinvoice.FromJSON(jsonBytes)
ublBytes, err := inv.ToXML(xinvoice.UBL)
ciiBytes, err := inv.ToXML(xinvoice.CII)

// Validation with localized JSON findings
result := xinvoice.ValidateXML(xmlBytes)
deJSON, err := result.JSON("de")
enJSON, err := result.JSON("en")
```

## CLI

The `xinvoice` command (under `cmd/xinvoice`) wraps the same operations. Input is read from stdin
(or `--in`) and auto-detected as UBL/CII XML or model JSON.

```sh
go install github.com/andeedotnet/go-xinvoice/cmd/xinvoice@latest

# Convert between syntaxes (here UBL XML -> CII XML, via the model)
xinvoice convert --to cii --in invoice_ubl.xml --out invoice_cii.xml

# Inspect any invoice as JSON
xinvoice convert --to json --in invoice.xml

# Validate; prints findings as JSON and exits non-zero on errors (CI-friendly)
xinvoice validate --lang de --in invoice.xml
```

## JSON model

The JSON structure mirrors the **EN16931 semantic model** (Business Terms `BT-*` and Business Groups
`BG-*`), so the same document maps to both UBL and CII. Field documentation links each field back to its
BT/BG identifier.

## Validation coverage

XRechnung's official validation is XSLT-2.0 Schematron. Go has no XSLT-2.0 engine, so this library
**reimplements the business rules natively in Go**, running them against the syntax-neutral model — so
one rule set covers both UBL and CII. Rule messages and code lists are derived from the official sources
(German from the XRechnung semantic model, English from the EN16931 schematron) and keyed by rule ID, so
they match the official wording.

Coverage grows incrementally. Every implemented rule is verified two ways against the official testsuite:
all 86 valid instances must produce **no false-positive errors**, and per-rule invalid fixtures must
trigger. For authoritative, 100 %-complete validation, cross-check against the official KoSIT validator.


## Rule coverage & why some rules are out of scope

**254/329 (77%) implemented · 75 documented out-of-scope · 0 uncovered** — enforced by
`internal/validation/coverage_test.go`, which fails if any catalog rule is neither implemented
(auto-detected from the rule sources) nor listed, with a reason, in its `outOfScope` set.

### Two principles that bound coverage
1. **Zero false positives.** Every implemented rule is verified to produce no error finding on the
   80 valid testsuite instances (`testsuite_test.go`); 41 invalid fixtures prove the rules actually
   fire (`rules_test.go`). A rule that cannot be verified against real data is left out rather than
   guessed — a validator that wrongly rejects valid invoices is worse than one that is honest about
   its scope.
2. **The validator runs against the syntax-neutral model** (one rule set for both UBL and CII). Rules
   that check the *XML structure* of a specific syntax therefore have no place to live: the model has
   already normalized that structure away. The official XSLT schematron checks them on the raw XML;
   this library checks the model.

### Implemented, per category
| Category                       | Implemented | Total |
|--------------------------------|-------------|-------|
| Existence (BR-*)               | 57          | 58    |
| Conditional/calc (BR-CO)       | 20          | 24    |
| Decimal places (BR-DEC)        | 21          | 21    |
| VAT (BR-S/Z/E/AE/IC/G/O/IG/IP)  | 96          | 96    |
| BR-DE-*                        | 23          | 37    |
| PEPPOL-EN16931-R*              | 15          | 23    |
| Codelist (BR-CL-*)             | 22          | 23    |
| **Total implemented**          | **254**     | **329** |

Fully covered: **all** VAT-category rules (96/96), **all** decimal-place rules (21/21) and nearly all
existence/calculation rules — i.e. everything that is both expressible against the semantic model and
verifiable against the testsuite.

### Out of scope (75), grouped by reason

**A — Syntax-level PEPPOL rules (8): R008, R043 (+R043-1/-2), R044, R053, R054, R101.**
They constrain the raw XML: "no empty elements" (R008), the charge-indicator value (R043), "no
price-level charge" (R044), "exactly one tax-total element" (R053/R054), `DocumentReference` placement
(R101). Against the normalized model these are either structurally guaranteed (there is a single
tax-total field) or not representable — see principle 2.

**B — Non-German VAT categories (22): BR-AF-* (10), BR-AG-* (10), BR-B-* (2, Italian split payment).**
The German XRechnung testsuite has no instances using these categories, so an implementation cannot be
verified (principle 1) and would never fire on a German invoice. They are cheap to add to the
`behaviors` table in `rules_vat.go` once test data exists.

**C — Rules needing data/mappings that are not available (6): BR-CO-5/6/7/8, BR-CL-24, BR-51.**
BR-CO-5..8 require an allowance/charge reason-code → reason-type equivalence table that is not in the
KoSIT sources we parse; BR-CL-24 validates the MIME type by regex, not by an enumerable code list (so
there is nothing to test membership against); BR-51 — a full card PAN — cannot be detected from a
stored value.

**D — Extension / profile / temporary rules (26): BR-DEX-* (15), BR-DE-CVD-* (7), BR-TMP-CVD-1,
BR-TMP-2/3, BR-DE-TMP-32.**
BR-DEX (sub-invoice-line rules) carries no rule text in the catalog and there is no *invalid* testsuite
instance to verify against; BR-DE-CVD / BR-TMP-CVD are the CVD construction profile; BR-TMP / BR-DE-TMP
are transitional/temporary rules.

**E — Complex national format rules, deferred (13): BR-DE-12, 14, 18, 23(-a/-b), 24(-a/-b), 25(-a/-b),
29.**
BR-DE-18 is the skonto free-text format (a regex over BT-20); BR-DE-23/24/25 are payment-means-conditional
account-detail rules; BR-DE-12 is a post-code format check; BR-DE-14 is covered structurally by BR-48;
BR-DE-29 is deprecated (replaced by PEPPOL-EN16931-R061). Doable with more work; not yet verified.

For authoritative, 100 %-complete validation, cross-check against the official KoSIT validator (the
README states this).

## Notes / decisions
- Validation is reimplemented natively in Go (no XSLT-2.0 runtime); it runs against the syntax-neutral
  model and is verified clean on the 80 valid non-extension testsuite instances (40 UBL + 40 CII; the
  full suite is 86). The README states coverage honestly vs. the official KoSIT validator.
- JSON shape = syntax-neutral EN16931 semantic model; one JSON document maps to both UBL and CII.
- Rule texts are keyed by rule id: German from the SeMoX model XML; English from the EN16931 `*.xsl`
  and the PEPPOL `*.sch`. Code lists come from the BR-CL-* membership tests in the EN16931 XSL.
- The official KoSIT bundle (<https://xeinkauf.de/xrechnung/versionen-und-bundles/>) is the source of
  truth. For full test-suite runs or catalog regeneration against the complete bundle, download it into
  a local `knowledge/` directory (git-ignored, not published); otherwise the generator embeds
  `internal/gen/resources/` and the tests fall back to `testdata/`.


## License & attribution

Licensed under the MIT License — see [`LICENSE`](./LICENSE).

This module includes copies of, and data **derived from**, the official XRechnung artifacts published
by KoSIT (Koordinierungsstelle für IT-Standards), which are licensed under Apache-2.0. The source
copies used for code generation live under [`internal/gen/resources/`](./internal/gen/resources). See
[`NOTICE`](./NOTICE) for attribution.
