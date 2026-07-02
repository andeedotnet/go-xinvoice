package model

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// sampleInvoice builds an invoice exercising scalars, optional groups, repeated
// groups, schemed identifiers, decimals, dates and an embedded attachment.
func sampleInvoice() *Invoice {
	return &Invoice{
		Number:                  "R2024-0001",
		IssueDate:               "2024-01-15",
		TypeCode:                "380",
		CurrencyCode:            "EUR",
		DueDate:                 "2024-02-15",
		BuyerReference:          "04011000-12345-39",
		SpecificationIdentifier: "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0",
		ProjectReference:        "PRJ-9",
		InvoicedObject:          &ID{Value: "OBJ-1", Scheme: "AAJ"},
		PaymentTerms:            "Zahlbar innerhalb von 30 Tagen",
		Notes: []Note{
			{SubjectCode: "AAI", Text: "Vielen Dank für Ihren Auftrag."},
			{Text: "Lieferung frei Haus."},
		},
		PrecedingInvoices: []PrecedingInvoiceReference{
			{Reference: "R2023-0099", IssueDate: "2023-12-01"},
		},
		Seller: Party{
			Name:        "Muster Lieferant GmbH",
			TradingName: "Muster",
			Identifiers: []ID{
				{Value: "549910"},
				{Value: "4000001000005", Scheme: "0088"},
			},
			LegalRegistrationID: &ID{Value: "HRB 12345", Scheme: "0198"},
			VATIdentifier:       "DE123456789",
			AdditionalLegalInfo: "Geschäftsführer: Max Muster",
			ElectronicAddress:   &ID{Value: "seller@example.de", Scheme: "EM"},
			Address: Address{
				Line1:       "Hauptstraße 1",
				City:        "Berlin",
				PostCode:    "10115",
				CountryCode: "DE",
			},
			Contact: &Contact{Point: "Vertrieb", Phone: "+49 30 1234567", Email: "vertrieb@example.de"},
		},
		Buyer: Party{
			Name:              "Beispiel Behörde",
			Identifiers:       []ID{{Value: "GOV-77", Scheme: "0204"}},
			ElectronicAddress: &ID{Value: "buyer@example.de", Scheme: "EM"},
			Address: Address{
				Line1:       "Amtsplatz 5",
				City:        "Hamburg",
				PostCode:    "20095",
				CountryCode: "DE",
			},
		},
		Payee: &Payee{Name: "Inkasso AG", Identifier: &ID{Value: "PAY-1", Scheme: "0088"}},
		Delivery: &Delivery{
			ActualDeliveryDate: "2024-01-10",
			InvoicingPeriod:    &Period{Start: "2024-01-01", End: "2024-01-31"},
			Address:            &Address{Line1: "Lager 3", City: "Hamburg", PostCode: "20095", CountryCode: "DE"},
		},
		PaymentInstructions: &PaymentInstructions{
			MeansTypeCode:         "58",
			RemittanceInformation: "R2024-0001",
			CreditTransfers: []CreditTransfer{
				{AccountIdentifier: "DE02120300000000202051", AccountName: "Muster Lieferant GmbH", ServiceProviderID: "BYLADEM1001"},
			},
		},
		Allowances: []AllowanceCharge{
			{Amount: "10.00", BaseAmount: "100.00", Percentage: "10", VATCategoryCode: "S", VATRate: "19", Reason: "Treuerabatt"},
		},
		Charges: []AllowanceCharge{
			{Amount: "5.00", VATCategoryCode: "S", VATRate: "19", Reason: "Versand"},
		},
		Totals: DocumentTotals{
			LineNetTotal:     "100.00",
			AllowanceTotal:   "10.00",
			ChargeTotal:      "5.00",
			TaxBasisTotal:    "95.00",
			TaxTotal:         "18.05",
			GrandTotal:       "113.05",
			DuePayableAmount: "113.05",
		},
		VATBreakdown: []VATBreakdown{
			{TaxableAmount: "95.00", TaxAmount: "18.05", CategoryCode: "S", Rate: "19"},
		},
		Documents: []SupportingDocument{
			{
				Reference:   "DOC-1",
				Description: "Lieferschein",
				Attachment:  &BinaryObject{MimeCode: "application/pdf", Filename: "ls.pdf", Content: []byte("%PDF-1.4 demo")},
			},
		},
		Lines: []Line{
			{
				ID:        "1",
				Quantity:  Quantity{Value: "10", Unit: "C62"},
				NetAmount: "100.00",
				Period:    &Period{Start: "2024-01-01", End: "2024-01-31"},
				Allowances: []LineAllowanceCharge{
					{Amount: "2.00", Reason: "Mengenrabatt"},
				},
				Price: Price{
					NetPrice:     "10.00",
					GrossPrice:   "12.00",
					BaseQuantity: &Quantity{Value: "1", Unit: "C62"},
				},
				VAT: LineVAT{CategoryCode: "S", Rate: "19"},
				Item: Item{
					Name:               "Musterartikel",
					Description:        "Ein Beispielartikel",
					SellerIdentifier:   "ART-100",
					StandardIdentifier: &ID{Value: "4012345678901", Scheme: "0160"},
					Classifications: []ItemClassification{
						{Code: "65434568", ListID: "STI", ListVersion: "2.1"},
					},
					CountryOfOrigin: "DE",
					Attributes: []ItemAttribute{
						{Name: "Farbe", Value: "blau"},
						{Name: "Gewicht", Value: "1.2 kg"},
					},
				},
			},
		},
	}
}

func TestJSONRoundTrip(t *testing.T) {
	orig := sampleInvoice()

	b, err := orig.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	got, err := FromJSON(b)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	if !reflect.DeepEqual(orig, got) {
		// Re-serialize both to surface the first differing line.
		b2, _ := got.ToJSON()
		t.Fatalf("round-trip mismatch.\n--- want ---\n%s\n--- got ---\n%s", b, b2)
	}
}

// TestJSONShape pins a few syntax decisions that the rest of the library and
// downstream consumers rely on.
func TestJSONShape(t *testing.T) {
	b, err := sampleInvoice().ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	s := string(b)

	// Decimals are bare JSON numbers, preserving trailing zeros.
	if !strings.Contains(s, `"grandTotal": 113.05`) {
		t.Errorf("amount not emitted as a bare number: %s", firstLineContaining(s, "grandTotal"))
	}
	if !strings.Contains(s, `"taxBasisTotal": 95.00`) {
		t.Errorf("trailing zeros not preserved: %s", firstLineContaining(s, "taxBasisTotal"))
	}
	// An ID without a scheme is a bare string; with a scheme it is an object.
	if !strings.Contains(s, `"value": "4000001000005"`) || !strings.Contains(s, `"scheme": "0088"`) {
		t.Errorf("schemed ID not emitted as object")
	}
	if !strings.Contains(s, `"reference": "R2023-0099"`) {
		t.Errorf("expected preceding invoice reference in output")
	}
	// Dates are quoted strings.
	if !strings.Contains(s, `"issueDate": "2024-01-15"`) {
		t.Errorf("date not emitted as YYYY-MM-DD string")
	}
	// omitempty drops unset optional fields.
	if strings.Contains(s, "taxRepresentative") {
		t.Errorf("unset optional group should be omitted")
	}
}

func TestIDUnionForms(t *testing.T) {
	// Bare string parses into Value with empty Scheme and re-emits as a string.
	var bare ID
	if err := json.Unmarshal([]byte(`"abc"`), &bare); err != nil {
		t.Fatalf("unmarshal bare: %v", err)
	}
	if bare.Value != "abc" || bare.Scheme != "" {
		t.Errorf("bare ID = %+v", bare)
	}
	if out, _ := json.Marshal(bare); string(out) != `"abc"` {
		t.Errorf("bare ID re-marshal = %s", out)
	}

	// Object form round-trips with the scheme.
	var obj ID
	if err := json.Unmarshal([]byte(`{"value":"x","scheme":"0088"}`), &obj); err != nil {
		t.Fatalf("unmarshal object: %v", err)
	}
	if obj.Value != "x" || obj.Scheme != "0088" {
		t.Errorf("object ID = %+v", obj)
	}
}

func TestDecimalValidation(t *testing.T) {
	for _, ok := range []string{"0", "100", "-5", "100.00", "-0.5", "1234567.89"} {
		if _, err := ParseDecimal(ok); err != nil {
			t.Errorf("ParseDecimal(%q) unexpected error: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "abc", "1.2.3", "1,5", "1e3", "+5", "."} {
		if _, err := ParseDecimal(bad); err == nil {
			t.Errorf("ParseDecimal(%q) expected error", bad)
		}
	}

	// Exact value access.
	r, ok := Decimal("19.99").Rat()
	if !ok || r.FloatString(2) != "19.99" {
		t.Errorf("Rat() = %v ok=%v", r, ok)
	}
}

// TestDecimalLengthGuard verifies that over-long and exponent-form decimals are
// rejected everywhere a Decimal enters arithmetic or parsing. Without the cap,
// a value with millions of digits (or a huge exponent that only the raw XML
// path can carry) drives big.Rat into super-linear time — a DoS reachable from
// any parsed document. See maxDecimalLen.
func TestDecimalLengthGuard(t *testing.T) {
	huge := strings.Repeat("9", 5_000_000)      // 5M digits, passes the shape regex
	exp := "1E1000000"                          // huge-exponent form (XML-only path)
	atCap := strings.Repeat("9", maxDecimalLen) // longest still-accepted value

	// ParseDecimal / JSON boundary must reject the over-long value.
	if _, err := ParseDecimal(huge); err == nil {
		t.Error("ParseDecimal accepted a 5M-digit value")
	}
	var d Decimal
	if err := json.Unmarshal([]byte(huge), &d); err == nil {
		t.Error("Decimal.UnmarshalJSON accepted a 5M-digit value")
	}

	// Rat() is the arithmetic choke point for both the JSON and the raw XML
	// paths; it must refuse both vectors without invoking big.Rat.SetString on
	// them (returning ok=false, which the validator treats as zero).
	for _, bad := range []string{huge, exp} {
		if _, ok := Decimal(bad).Rat(); ok {
			t.Errorf("Rat() accepted pathological value %.12q…", bad)
		}
	}

	// A value exactly at the cap is still valid (no legitimate value is that long).
	if _, err := ParseDecimal(atCap); err != nil {
		t.Errorf("ParseDecimal rejected a value at the length cap: %v", err)
	}
}

func TestDateValidation(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte(`"2024-13-01"`), &d); err == nil {
		t.Errorf("expected invalid month to fail")
	}
	if err := json.Unmarshal([]byte(`"2024-02-29"`), &d); err != nil {
		t.Errorf("valid leap date failed: %v", err)
	}
}

func firstLineContaining(s, sub string) string {
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, sub) {
			return strings.TrimSpace(line)
		}
	}
	return "(not found)"
}
