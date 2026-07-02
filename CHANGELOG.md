# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.4] - 2026-07-02

### Added
- Fuzz tests for the XML and JSON parsers (`FuzzParseXML`, `FuzzFromJSON`),
  seeded from the committed testsuite instances. CI now also runs `govulncheck`,
  a short live fuzz-smoke pass over the parsers, and reports coverage.
- Negative unit tests for the BR-DE-27 / BR-DE-28 seller-contact plausibility
  warnings (previously untested), plus a focused `validEmail` test.

### Changed
- Tightened the BR-DE-28 e-mail plausibility check to reject clearly-malformed
  addresses such as `a@b.` and `a@.b` (empty domain label around the dot).
- `periodOK` now compares dates chronologically (parsed via `Date.Time()`) with a
  lexical fallback, instead of relying solely on ISO string order.
- Clarified the out-of-scope reason for BR-DE-12 (it exists only in the SeMoX
  model with no schematron assertion and is subsumed by BR-DE-11 on BT-78, so it
  cannot be verified independently).

### Removed
- Dead `categoriesUsed` helper in the VAT rules (never called).

## [0.1.3] - 2026-07-02

### Security
- Bounded the length of `Decimal` values (max 40 characters — real EN16931
  amounts and quantities are ≤ 21). Without the cap, a field carrying millions
  of digits, or a huge-exponent form such as `1E1000000` that the raw XML path
  could carry, drove the validator's `big.Rat` arithmetic into super-linear
  time (a single ~15 MB decimal took ~2 minutes) — a denial-of-service vector
  reachable from any parsed document via `Validate` / `ValidateXML`. The cap is
  enforced at the parse boundaries (`ParseDecimal`, JSON marshal/unmarshal) and,
  crucially, inside `Decimal.Rat()` — the single choke point every arithmetic
  rule passes through — so both the JSON and the XML code paths are covered.
  Over-long or exponent values are now rejected (treated as zero by the
  validator) instead of being multiplied.

## [0.1.2] - 2026-06-15

### Added
- Curated English texts for the German national rules (`BR-DE-*`), whose official
  rule texts exist only in German. `ValidationResult.JSON("en")` now renders these
  rules in English instead of falling back to the German text.

## [0.1.1] - 2026-06-15

### Added
- Validation findings now include a `locationLabel` field in the JSON output — the
  localized BT/BG term label for the finding's location (e.g. `BT-10` →
  "Käuferreferenz" / "Buyer reference"), so callers can show what the field is.

### Changed
- Module now targets Go 1.26.

## [0.1.0] - 2026-06-12

Initial release.

### Added
- Syntax-neutral EN16931 / XRechnung 3.0.2 semantic model (`model`), mapping one
  JSON document to either official XML syntax.
- UBL 2.1 and UN/CEFACT CII D16B read + write with auto-detection on parse.
- Native Go reimplementation of the EN16931 / XRechnung business rules
  (254/329 rules; see `TODO.md`), running against the model so one rule set
  covers both syntaxes. Findings render as JSON in German or English.
- `xinvoice` CLI: `convert` between JSON/UBL/CII and `validate` (exits non-zero
  on error findings, so it can gate a pipeline).
- Check-digit helpers for GLN, IBAN and generic mod-11.
- Embedded rule/label/code-list catalog generated from the official KoSIT
  sources (`go generate ./internal/catalog`).

[0.1.2]: https://github.com/andeedotnet/go-xinvoice/releases/tag/v0.1.2
[0.1.1]: https://github.com/andeedotnet/go-xinvoice/releases/tag/v0.1.1
[0.1.0]: https://github.com/andeedotnet/go-xinvoice/releases/tag/v0.1.0
