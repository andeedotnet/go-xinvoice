# TODO / Status

Status of **go-xinvoice** against XRechnung 3.0.2 (EN16931 CIUS + Extension).
The README carries the narrative; this file tracks the per-category state and the
concrete next steps. Rule coverage is enforced by
`internal/validation/coverage_test.go` (it fails if a catalog rule is neither
implemented nor listed, with a reason, in its `outOfScope` set).

## Conversion (JSON ⇄ UBL ⇄ CII)

- [x] Syntax-neutral EN16931 semantic model (BT-*/BG-*) — `model/`
- [x] UBL 2.1 read + write — `internal/ubl`
- [x] UN/CEFACT CII D16B read + write — `internal/cii`
- [x] Auto-detection of UBL vs CII on parse
- [x] Cross-syntax round-trip verified — `internal/cii/equivalence_test.go`
- [ ] Extension / sub-invoice-line (BG-DEX) modelling — currently skipped on parse

## Validation coverage

**254 / 329 rules implemented · 75 documented out-of-scope · 0 uncovered.**

| Category                        | Implemented | Total |
|---------------------------------|-------------|-------|
| Existence (BR-*)                | 57          | 58    |
| Conditional/calc (BR-CO)        | 20          | 24    |
| Decimal places (BR-DEC)         | 21          | 21    |
| VAT (BR-S/Z/E/AE/IC/G/O/IG/IP)  | 96          | 96    |
| BR-DE-*                         | 23          | 37    |
| PEPPOL-EN16931-R*               | 15          | 23    |
| Codelist (BR-CL-*)              | 22          | 23    |

Two invariants bound this (see README): **zero false positives** (a rule ships
only if verified against real testsuite data) and **the validator runs on the
normalized model, not raw XML**.

## Next up (implementable — out of scope today only for want of work or test data)

- [ ] **Non-German VAT categories** `BR-AF-*`, `BR-AG-*` — add to the `behaviors`
      table in `internal/validation/rules_vat.go` once test instances exist.
- [ ] **National format rules** `BR-DE-12` (post-code), `BR-DE-18` (skonto
      free-text regex over BT-20), `BR-DE-23/24/25` (payment-means-conditional
      account details).
- [ ] `BR-CO-5/6/7/8` — need an allowance/charge reason-code ↔ reason-type
      mapping table that is not present in the parsed KoSIT sources.

## Out of scope by design (won't implement)

Rules that constrain raw per-syntax XML structure have no place in a validator
that runs on the normalized model: syntax-level PEPPOL rules (R008, R043, R044,
R053/R054, R101) and rules the model structurally guarantees. Transitional /
profile rules (`BR-TMP-*`, `BR-DE-CVD-*`, `BR-DEX-*`) and Italian split-payment
(`BR-B-*`) are likewise excluded. See `outOfScope` in `coverage_test.go` for the
full list with reasons.

For authoritative, 100 %-complete validation, cross-check against the official
KoSIT validator.
