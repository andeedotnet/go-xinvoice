package catalog

// brDEEnglish provides curated English translations for the German national
// XRechnung rules (BR-DE-*), whose official rule texts exist only in German.
// They are used as a fallback by [Rule.Text] when "en" is requested and the
// generated catalog has no English text. These translations are maintained by
// hand (they are not derivable from the upstream sources) and are intended to be
// faithful, readable renderings of the German originals — not official text.
var brDEEnglish = map[string]string{
	"BR-DE-1":    `An Invoice (INVOICE) must contain information from group BG-16 (SELLER CONTACT).`,
	"BR-DE-2":    `Group BG-6 (SELLER CONTACT) must be provided.`,
	"BR-DE-3":    `Element BT-37 must be provided.`,
	"BR-DE-4":    `Element BT-38 must be provided.`,
	"BR-DE-5":    `Element BT-41 must be provided.`,
	"BR-DE-6":    `Element BT-42 must be provided.`,
	"BR-DE-7":    `Element BT-43 must be provided.`,
	"BR-DE-8":    `Element BT-52 must be provided.`,
	"BR-DE-9":    `Element BT-53 must be provided.`,
	"BR-DE-10":   `Element BT-77 must be provided when group BG-15 is provided.`,
	"BR-DE-11":   `Element BT-78 must be provided when group BG-15 is provided.`,
	"BR-DE-12":   `A post code must be provided in element BT-78.`,
	"BR-DE-14":   `Element BT-119 must be provided.`,
	"BR-DE-15":   `Element BT-10 must be provided.`,
	"BR-DE-16":   `If the tax category codes S, Z, E, AE, K, G, L or M are used in an invoice, at least one of the elements BT-31, BT-32 or BG-11 must be provided.`,
	"BR-DE-17":   `Element BT-3 should use one of the following codes from code list UNTDID 1001: 326 (Partial invoice), 380 (Commercial invoice), 384 (Corrected invoice), 389 (Self-billed invoice), 381 (Credit note), 875 (Partial construction invoice), 876 (Partial final construction invoice), 877 (Final construction invoice).`,
	"BR-DE-18":   `Cash-discount (Skonto) information must be provided in element BT-20 in the prescribed XRechnung format (segments SKONTO, TAGE=n, PROZENT=n, optionally BASISBETRAG=n; uppercase, '#'-separated).`,
	"BR-DE-19":   `BT-84 should contain a correct IBAN when payment means code 58 (SEPA credit transfer) is requested in BT-81.`,
	"BR-DE-20":   `BT-91 should contain a correct IBAN when payment means code 59 (SEPA direct debit) is requested in BT-81.`,
	"BR-DE-21":   `Element BT-24 should syntactically match the specification identifier of the XRechnung standard.`,
	"BR-DE-22":   `Documents attached to a submitted invoice in BG-24 must have a unique file name (case-insensitive) in element BT-125.`,
	"BR-DE-23":   `If BT-81 contains a credit-transfer code (30, 58), BG-17 must be provided; BG-18 and BG-19 must not be provided in that case.`,
	"BR-DE-23-a": `If BT-81 "Payment means type code" contains a credit-transfer code (30, 58), BG-17 "CREDIT TRANSFER" must be provided.`,
	"BR-DE-23-b": `If BT-81 "Payment means type code" contains a credit-transfer code (30, 58), BG-18 and BG-19 must not be provided.`,
	"BR-DE-24":   `If BT-81 contains a payment-card code (48, 54, 55), exactly BG-18 must be provided; BG-17 and BG-19 must not be provided in that case.`,
	"BR-DE-24-a": `If BT-81 "Payment means type code" contains a payment-card code (48, 54, 55), exactly BG-18 "PAYMENT CARD INFORMATION" must be provided.`,
	"BR-DE-24-b": `If BT-81 "Payment means type code" contains a payment-card code (48, 54, 55), BG-17 and BG-19 must not be provided.`,
	"BR-DE-25":   `If BT-81 contains a direct-debit code (59), exactly BG-19 must be provided; BG-17 and BG-18 must not be provided in that case.`,
	"BR-DE-25-a": `If BT-81 "Payment means type code" contains a direct-debit code (59), exactly BG-19 "DIRECT DEBIT" must be provided.`,
	"BR-DE-25-b": `If BT-81 "Payment means type code" contains a direct-debit code (59), BG-17 and BG-18 must not be provided.`,
	"BR-DE-26":   `When code 384 (Corrected invoice) is used in element BT-3, BG-3 should be present at least once.`,
	"BR-DE-27":   `Element BT-42 should contain a valid telephone number. A valid telephone number must contain at least three digits.`,
	"BR-DE-28":   `Element BT-43 should contain a valid e-mail address. Note: technical implementation details are in the XRechnung Schematron rule BR-DE-28.`,
	"BR-DE-29":   `This rule has been replaced by PEPPOL-EN16931-R061 since XRechnung 3.0.0.`,
	"BR-DE-30":   `Element BT-90 must be provided when group BG-19 is provided.`,
	"BR-DE-31":   `Element BT-91 must be provided when group BG-19 is provided.`,

	"BR-DE-CVD-1":   `The "Contract reference" element (BT-12) must be provided.`,
	"BR-DE-CVD-2":   `The "Tender or lot reference" element (BT-17) must be provided.`,
	"BR-DE-CVD-3":   `An invoice must contain at least one INVOICE LINE (BG-25) in which the scheme identifier of "Item classification identifier" (BT-158) is 'CVD' and the "Item attribute name" (BT-160) is 'cva'.`,
	"BR-DE-CVD-4":   `An "Item classification identifier" (BT-158) with scheme identifier 'CVD' must contain a value from the list of permitted vehicle categories.`,
	"BR-DE-CVD-5":   `When the "Item attribute name" (BT-160) within ITEM ATTRIBUTES (BG-32) is 'cva', the "Item attribute value" (BT-161) must contain one of the permitted values.`,
	"BR-DE-CVD-6-a": `When the scheme identifier of "Item classification identifier" (BT-158) is 'CVD', exactly one "Item attribute name" (BT-160) with the value 'cva' must be present in the same invoice line.`,
	"BR-DE-CVD-6-b": `When "Item attribute name" (BT-160) is 'cva', exactly one "Item classification identifier" (BT-158) with scheme identifier 'CVD' must be present in the same invoice line.`,

	"BR-DE-TMP-32": `To state the delivery/performance date, an invoice should contain either BT-72 "Actual delivery date", BG-14 "Invoicing period", or BG-26 "Invoice line period" in each invoice line.`,
}
