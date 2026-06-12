# Test instances

A curated subset of the official KoSIT XRechnung 3.0.2 **test suite** instances, in both UBL
(`*_ubl.xml`) and CII (`*_uncefact.xml`) syntaxes. They let the round-trip, validation and
cross-syntax tests run without the full upstream bundle (which is not part of the published
module; download it from <https://xeinkauf.de/xrechnung/versionen-und-bundles/>).

The selection covers the meaningful cases: standard invoices (VAT categories S / O / AE),
the comprehensive test cases, a CVD construction-profile case, and the UBL-only extensions
(sub-invoice lines, third-party payments).

These files are licensed under the **Apache License 2.0** by KoSIT; see the repository `NOTICE`.
The tests prefer a full upstream test suite (dropped into `knowledge/`) when present and fall
back to this directory otherwise.
