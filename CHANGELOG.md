# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
