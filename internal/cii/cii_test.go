package cii

import (
	"reflect"
	"strings"
	"testing"

	"github.com/andeedotnet/go-xinvoice/model"
)

// richInvoice exercises the CII-mapped fields (including ones UBL cannot carry,
// such as note subject codes BT-21 and the direct-debit creditor id BT-90). It
// avoids the few deferred fields (BT-7/BT-8 tax point date/code).
func richInvoice() *model.Invoice {
	return &model.Invoice{
		Number:                   "R-2024-7",
		IssueDate:                "2024-03-01",
		DueDate:                  "2024-03-31",
		TypeCode:                 "380",
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
			MeansTypeCode:         "59",
			MeansText:             "SEPA Lastschrift",
			RemittanceInformation: "R-2024-7",
			DirectDebit:           &model.DirectDebit{MandateReference: "MANDATE-1", CreditorIdentifier: "DE98ZZZ09999999999", DebitedAccount: "DE02120300000000202051"},
		},
		Allowances: []model.AllowanceCharge{{Amount: "10.00", BaseAmount: "100.00", Percentage: "10", VATCategoryCode: "S", VATRate: "19", Reason: "Rabatt", ReasonCode: "95"}},
		Charges:    []model.AllowanceCharge{{Amount: "5.00", VATCategoryCode: "S", VATRate: "19", Reason: "Versand"}},
		Totals: model.DocumentTotals{
			LineNetTotal: "100.00", AllowanceTotal: "10.00", ChargeTotal: "5.00",
			TaxBasisTotal: "95.00", TaxTotal: "18.05", TaxTotalAccountingCurrency: "18.05",
			GrandTotal: "113.05", PaidAmount: "13.05", RoundingAmount: "0.00", DuePayableAmount: "100.00",
		},
		VATBreakdown: []model.VATBreakdown{{TaxableAmount: "95.00", TaxAmount: "18.05", CategoryCode: "S", Rate: "19", ExemptionReasonText: "n/a", ExemptionReasonCode: "VATEX-EU-79-C"}},
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

func TestSyntheticRoundTrip(t *testing.T) {
	orig := richInvoice()
	xmlBytes, err := Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := Parse(xmlBytes)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		j1, _ := orig.ToJSON()
		j2, _ := got.ToJSON()
		t.Errorf("model -> CII -> model lost data.\n%s", firstDiff(string(j1), string(j2)))
	}
}

func TestMarshalProducesCIINamespace(t *testing.T) {
	xmlBytes, err := Marshal(richInvoice())
	if err != nil {
		t.Fatal(err)
	}
	if !Detect(xmlBytes) {
		t.Error("Marshal output is not detected as CII")
	}
	s := string(xmlBytes)
	if !strings.Contains(s, "<rsm:CrossIndustryInvoice") || !strings.Contains(s, "xmlns:ram=") || !strings.Contains(s, "<ram:ID>") {
		t.Errorf("expected conventional rsm:/ram: prefixes:\n%.300s", s)
	}
}
