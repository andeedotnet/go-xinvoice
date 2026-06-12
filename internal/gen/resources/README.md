# Generator resources

Verbatim copies of the official KoSIT XRechnung 3.0.2 source artifacts that the
catalog generator (`internal/gen`) reads. They are embedded (`go:embed`) so the
generator runs without a copy of the full upstream bundle present. The bundle is
downloaded from <https://xeinkauf.de/xrechnung/versionen-und-bundles/>.

These files are licensed under the **Apache License 2.0** by KoSIT
(Koordinierungsstelle für IT-Standards); see the repository `NOTICE`.

| File | Copied from (in the KoSIT bundle) | Used for |
|------|-----------------------------------|----------|
| `xrechnung-cius-model.xml` | `…-xrechnung-model-…/model/xrechnung-cius-model.xml` | German rule texts + BT/BG terms per rule |
| `EN16931-UBL-validation.xsl` | `…-validator-configuration-…/EN16931-UBL-validation.xsl` | English EN16931 rule texts + severity, BR-CL code lists (UBL) |
| `EN16931-CII-validation.xsl` | `…-validator-configuration-…/EN16931-CII-validation.xsl` | English EN16931 rule texts + severity, BR-CL code lists (CII) |
| `XRechnung-UBL-validation.sch` | `…-schematron-…/schematron/ubl/XRechnung-UBL-validation.sch` | BR-DE (German) + PEPPOL (English) rules (UBL) |
| `XRechnung-CII-validation.sch` | `…-schematron-…/schematron/cii/XRechnung-CII-validation.sch` | BR-DE (German) + PEPPOL (English) rules (CII) |
| `l10n-de.xml` | `…-visualization-…/xsl/l10n/de.xml` | German BT/BG labels |
| `l10n-en.xml` | `…-visualization-…/xsl/l10n/en.xml` | English BT/BG labels |

## Refreshing after a bundle update

Download a newer KoSIT bundle from
<https://xeinkauf.de/xrechnung/versionen-und-bundles/>, re-copy the seven files
into this directory (keeping the names above), then run `go generate ./...` and
review the diff in `internal/catalog/*_gen.go`.
