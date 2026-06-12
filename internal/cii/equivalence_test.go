package cii_test

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/andeedotnet/go-xinvoice/internal/cii"
	"github.com/andeedotnet/go-xinvoice/internal/ubl"
	"github.com/andeedotnet/go-xinvoice/model"
)

// TestCrossSyntaxEquivalence proves the model is syntax-neutral: one source
// invoice, serialized to both UBL and CII and parsed back, yields the same model
// through either syntax (and matches the source). Decimals are compared
// numerically.
//
// (The official testsuite cannot be used for this directly: its UBL and CII
// files for the same number carry different placeholder data, so they are not
// content-identical translations.)
func TestCrossSyntaxEquivalence(t *testing.T) {
	source := sourceInvoice()

	ublXML, err := ubl.Marshal(source)
	if err != nil {
		t.Fatalf("ubl.Marshal: %v", err)
	}
	ciiXML, err := cii.Marshal(source)
	if err != nil {
		t.Fatalf("cii.Marshal: %v", err)
	}

	viaUBL, err := ubl.Parse(ublXML)
	if err != nil {
		t.Fatalf("ubl.Parse: %v", err)
	}
	viaCII, err := cii.Parse(ciiXML)
	if err != nil {
		t.Fatalf("cii.Parse: %v", err)
	}

	canonicalize(source)
	canonicalize(viaUBL)
	canonicalize(viaCII)

	if !reflect.DeepEqual(viaUBL, viaCII) {
		ju, _ := viaUBL.ToJSON()
		jc, _ := viaCII.ToJSON()
		t.Fatalf("UBL and CII models differ:\n%s", firstLineDiff(string(ju), string(jc)))
	}
	if !reflect.DeepEqual(source, viaUBL) {
		js, _ := source.ToJSON()
		ju, _ := viaUBL.ToJSON()
		t.Errorf("round-trip changed the model:\n%s", firstLineDiff(string(js), string(ju)))
	}
}

// sourceInvoice is a comprehensive CIUS invoice that both syntaxes can fully
// represent (no UBL-only extension fields such as sub-lines or third-party
// payments).
func sourceInvoice() *model.Invoice {
	return &model.Invoice{
		Number:                   "R-2024-7",
		IssueDate:                "2024-03-01",
		DueDate:                  "2024-03-31",
		TypeCode:                 "380",
		TaxPointDate:             "2024-02-28",
		CurrencyCode:             "EUR",
		VATAccountingCurrency:    "EUR",
		BuyerReference:           "04011000-12345-39",
		BuyerAccountingReference: "CC-9",
		SpecificationIdentifier:  "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0",
		BusinessProcessType:      "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
		PurchaseOrderReference:   "PO-1",
		SalesOrderReference:      "SO-1",
		ContractReference:        "K-1",
		ProjectReference:         "PRJ-1",
		DespatchAdviceReference:  "DA-1",
		ReceivingAdviceReference: "RA-1",
		TenderOrLotReference:     "TL-1",
		PaymentTerms:             "Zahlbar in 30 Tagen",
		Notes:                    []model.Note{{SubjectCode: "AAI", Text: "Vielen Dank"}, {Text: "Lieferung frei Haus"}},
		InvoicedObject:           &model.ID{Value: "OBJ-9", Scheme: "AAJ"},
		PrecedingInvoices:        []model.PrecedingInvoiceReference{{Reference: "R-2023-9", IssueDate: "2023-12-01"}},
		Seller: model.Party{
			Name:                "Lieferant GmbH",
			TradingName:         "Lieferant",
			Identifiers:         []model.ID{{Value: "549910"}, {Value: "4000001000005", Scheme: "0088"}},
			LegalRegistrationID: &model.ID{Value: "HRB123", Scheme: "0198"},
			VATIdentifier:       "DE123456789",
			TaxRegistrationID:   "201/123/40000",
			AdditionalLegalInfo: "Geschäftsführer: Max Muster",
			ElectronicAddress:   &model.ID{Value: "seller@example.de", Scheme: "EM"},
			Address:             model.Address{Line1: "Hauptstr. 1", Line2: "Geb. 2", Line3: "links", City: "Berlin", PostCode: "10115", Subdivision: "BE", CountryCode: "DE"},
			Contact:             &model.Contact{Point: "Vertrieb", Phone: "+49301234567", Email: "v@example.de"},
		},
		Buyer: model.Party{
			Name:              "Behörde",
			Identifiers:       []model.ID{{Value: "GOV-7", Scheme: "0204"}},
			VATIdentifier:     "DE987654321",
			ElectronicAddress: &model.ID{Value: "buyer@example.de", Scheme: "EM"},
			Address:           model.Address{Line1: "Amtsplatz 5", City: "Hamburg", PostCode: "20095", CountryCode: "DE"},
			Contact:           &model.Contact{Point: "Einkauf"},
		},
		Payee:             &model.Payee{Name: "Inkasso AG", Identifier: &model.ID{Value: "PAY-1", Scheme: "0088"}, LegalRegistrationID: &model.ID{Value: "HRB999", Scheme: "0198"}},
		TaxRepresentative: &model.TaxRepresentative{Name: "Steuer Vertreter", VATIdentifier: "DE999999999", Address: model.Address{Line1: "Steuerweg 1", City: "Köln", PostCode: "50667", CountryCode: "DE"}},
		Delivery: &model.Delivery{
			PartyName:          "Lager Nord",
			LocationIdentifier: &model.ID{Value: "LOC-1", Scheme: "0088"},
			ActualDeliveryDate: "2024-02-25",
			InvoicingPeriod:    &model.Period{Start: "2024-02-01", End: "2024-02-29"},
			Address:            &model.Address{Line1: "Lagerstr. 3", City: "Hamburg", PostCode: "20095", CountryCode: "DE"},
		},
		PaymentInstructions: &model.PaymentInstructions{
			MeansTypeCode:         "58",
			MeansText:             "SEPA",
			RemittanceInformation: "R-2024-7",
			CreditTransfers: []model.CreditTransfer{
				{AccountIdentifier: "DE02120300000000202051", AccountName: "Lieferant GmbH", ServiceProviderID: "BYLADEM1001"},
			},
		},
		Allowances: []model.AllowanceCharge{{Amount: "10.00", BaseAmount: "100.00", Percentage: "10", VATCategoryCode: "S", VATRate: "19", Reason: "Rabatt", ReasonCode: "95"}},
		Charges:    []model.AllowanceCharge{{Amount: "5.00", VATCategoryCode: "S", VATRate: "19", Reason: "Versand"}},
		Totals: model.DocumentTotals{
			LineNetTotal: "100.00", AllowanceTotal: "10.00", ChargeTotal: "5.00",
			TaxBasisTotal: "95.00", TaxTotal: "18.05", TaxTotalAccountingCurrency: "18.05",
			GrandTotal: "113.05", PaidAmount: "13.05", RoundingAmount: "0.00", DuePayableAmount: "100.00",
		},
		VATBreakdown: []model.VATBreakdown{{TaxableAmount: "95.00", TaxAmount: "18.05", CategoryCode: "S", Rate: "19"}},
		Documents: []model.SupportingDocument{
			{Reference: "DOC-1", Description: "Lieferschein", Attachment: &model.BinaryObject{MimeCode: "application/pdf", Filename: "ls.pdf", Content: []byte("PDFDATA")}},
			{Reference: "DOC-2", Description: "Extern", ExternalLocation: "https://example.de/doc"},
		},
		Lines: []model.Line{
			{
				ID: "1", Note: "Pos 1",
				Quantity: model.Quantity{Value: "10", Unit: "C62"}, NetAmount: "100.00",
				OrderLineReference: "POL-1", BuyerAccountingReference: "CC-L1",
				ObjectIdentifier: &model.ID{Value: "LOBJ-1", Scheme: "AAJ"},
				Period:           &model.Period{Start: "2024-02-01", End: "2024-02-29"},
				Allowances:       []model.LineAllowanceCharge{{Amount: "2.00", BaseAmount: "20.00", Percentage: "10", Reason: "Mengenrabatt", ReasonCode: "95"}},
				Charges:          []model.LineAllowanceCharge{{Amount: "1.00", Reason: "Eil"}},
				Price:            model.Price{NetPrice: "10.00", Discount: "2.00", GrossPrice: "12.00", BaseQuantity: &model.Quantity{Value: "1", Unit: "C62"}},
				VAT:              model.LineVAT{CategoryCode: "S", Rate: "19"},
				Item: model.Item{
					Name: "Artikel", Description: "Ein Artikel",
					SellerIdentifier: "ART-1", BuyerIdentifier: "BART-1",
					StandardIdentifier: &model.ID{Value: "4012345678901", Scheme: "0160"},
					Classifications:    []model.ItemClassification{{Code: "65434568", ListID: "STI", ListVersion: "2.1"}},
					CountryOfOrigin:    "DE",
					Attributes:         []model.ItemAttribute{{Name: "Farbe", Value: "blau"}},
				},
			},
		},
	}
}

var decimalType = reflect.TypeOf(model.Decimal(""))

// canonicalize rewrites every Decimal field to its exact rational form, so that
// numerically-equal amounts (12.6 vs 12.60) compare equal.
func canonicalize(v any) { canonValue(reflect.ValueOf(v)) }

func canonValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			canonValue(v.Elem())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			canonValue(v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if f.Type() == decimalType {
				if f.CanSet() && f.String() != "" {
					if r, ok := new(big.Rat).SetString(f.String()); ok {
						f.SetString(r.RatString())
					}
				}
				continue
			}
			canonValue(f)
		}
	}
}

func firstLineDiff(a, b string) string {
	la, lb := splitLines(a), splitLines(b)
	for i := 0; i < len(la) || i < len(lb); i++ {
		var x, y string
		if i < len(la) {
			x = la[i]
		}
		if i < len(lb) {
			y = lb[i]
		}
		if x != y {
			return "  ubl: " + x + "\n  cii: " + y
		}
	}
	return "(structural difference)"
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, trimSpace(s[start:i]))
			start = i + 1
		}
	}
	out = append(out, trimSpace(s[start:]))
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}
