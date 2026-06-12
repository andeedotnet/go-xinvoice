package validation

import (
	"strings"
	"testing"

	"github.com/andeedotnet/go-xinvoice/model"
)

// validInvoice is a minimal invoice that passes every enforced rule. Each
// negative test mutates a copy of it.
func validInvoice() *model.Invoice {
	return &model.Invoice{
		SpecificationIdentifier: "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0",
		BusinessProcessType:     "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
		Number:                  "R1",
		IssueDate:               "2024-01-01",
		TypeCode:                "380",
		CurrencyCode:            "EUR",
		DueDate:                 "2024-02-15",
		BuyerReference:          "04011000-12345-39",
		Seller: model.Party{
			Name:              "Seller GmbH",
			VATIdentifier:     "DE123456789",
			ElectronicAddress: &model.ID{Value: "s@example.de", Scheme: "EM"},
			Address:           model.Address{City: "Berlin", PostCode: "10115", CountryCode: "DE"},
			Contact:           &model.Contact{Point: "Sales", Phone: "+4930123", Email: "s@example.de"},
		},
		Buyer: model.Party{
			Name:              "Buyer AG",
			ElectronicAddress: &model.ID{Value: "b@example.de", Scheme: "EM"},
			Address:           model.Address{City: "Hamburg", PostCode: "20095", CountryCode: "DE"},
		},
		PaymentInstructions: &model.PaymentInstructions{MeansTypeCode: "58", CreditTransfers: []model.CreditTransfer{{AccountIdentifier: "DE02120300000000202051"}}},
		Totals: model.DocumentTotals{
			LineNetTotal: "100.00", TaxBasisTotal: "100.00", TaxTotal: "19.00",
			GrandTotal: "119.00", DuePayableAmount: "119.00",
		},
		VATBreakdown: []model.VATBreakdown{{TaxableAmount: "100.00", TaxAmount: "19.00", CategoryCode: "S", Rate: "19"}},
		Lines: []model.Line{{
			ID:        "1",
			Quantity:  model.Quantity{Value: "1", Unit: "C62"},
			NetAmount: "100.00",
			Price:     model.Price{NetPrice: "100.00"},
			VAT:       model.LineVAT{CategoryCode: "S", Rate: "19"},
			Item:      model.Item{Name: "Widget"},
		}},
	}
}

// fires reports whether the given rule appears among the findings.
func fires(inv *model.Invoice, rule string) bool {
	for _, f := range Validate(inv).Findings {
		if f.Rule == rule {
			return true
		}
	}
	return false
}

func TestValidInvoiceHasNoErrors(t *testing.T) {
	res := Validate(validInvoice())
	if !res.Valid() {
		var got []string
		for _, f := range res.Findings {
			if f.Severity == "error" {
				got = append(got, f.Rule)
			}
		}
		t.Fatalf("valid invoice produced errors: %v", got)
	}
}

func TestRulesFireOnViolations(t *testing.T) {
	cases := []struct {
		name string
		rule string
		mut  func(*model.Invoice)
	}{
		{"missing number", "BR-2", func(i *model.Invoice) { i.Number = "" }},
		{"missing issue date", "BR-3", func(i *model.Invoice) { i.IssueDate = "" }},
		{"missing currency", "BR-5", func(i *model.Invoice) { i.CurrencyCode = "" }},
		{"missing seller name", "BR-6", func(i *model.Invoice) { i.Seller.Name = "" }},
		{"missing seller country", "BR-9", func(i *model.Invoice) { i.Seller.Address.CountryCode = "" }},
		{"no lines", "BR-16", func(i *model.Invoice) { i.Lines = nil }},
		{"missing line item name", "BR-25", func(i *model.Invoice) { i.Lines[0].Item.Name = "" }},
		{"invalid currency code", "BR-CL-04", func(i *model.Invoice) { i.CurrencyCode = "ZZZ" }},
		{"invalid country code", "BR-CL-14", func(i *model.Invoice) { i.Seller.Address.CountryCode = "ZZ" }},
		{"invalid VAT category", "BR-CL-17", func(i *model.Invoice) { i.VATBreakdown[0].CategoryCode = "Q" }},
		{"invalid unit code", "BR-CL-23", func(i *model.Invoice) { i.Lines[0].Quantity.Unit = "ZZZZ" }},
		{"line net sum mismatch", "BR-CO-10", func(i *model.Invoice) { i.Lines[0].NetAmount = "90.00" }},
		{"grand total mismatch", "BR-CO-15", func(i *model.Invoice) { i.Totals.GrandTotal = "200.00" }},
		{"category VAT mismatch", "BR-CO-17", func(i *model.Invoice) { i.VATBreakdown[0].TaxAmount = "5.00" }},
		{"missing buyer reference", "BR-DE-15", func(i *model.Invoice) { i.BuyerReference = "" }},
		{"missing payment instructions", "BR-DE-1", func(i *model.Invoice) { i.PaymentInstructions = nil }},
		{"missing seller contact", "BR-DE-2", func(i *model.Invoice) { i.Seller.Contact = nil }},
		{"missing business process", "PEPPOL-EN16931-R001", func(i *model.Invoice) { i.BusinessProcessType = "" }},
		{"missing seller e-address", "PEPPOL-EN16931-R020", func(i *model.Invoice) { i.Seller.ElectronicAddress = nil }},

		// VAT category rules.
		{"standard rate must be positive", "BR-S-5", func(i *model.Invoice) { i.Lines[0].VAT.Rate = "0" }},
		{"standard breakdown with exemption", "BR-S-10", func(i *model.Invoice) { i.VATBreakdown[0].ExemptionReasonText = "x" }},
		{"category taxable basis mismatch", "BR-S-8", func(i *model.Invoice) { i.Lines[0].NetAmount = "50.00" }},
		{"used category missing from breakdown", "BR-E-1", func(i *model.Invoice) {
			i.Lines[0].VAT = model.LineVAT{CategoryCode: "E", Rate: "0"}
		}},
		{"exempt breakdown without reason", "BR-E-10", func(i *model.Invoice) {
			i.Lines[0].VAT = model.LineVAT{CategoryCode: "E", Rate: "0"}
			i.VATBreakdown[0] = model.VATBreakdown{TaxableAmount: "100.00", TaxAmount: "0.00", CategoryCode: "E", Rate: "0"}
		}},
		{"not-subject-to-VAT with VAT id", "BR-O-2", func(i *model.Invoice) {
			i.Lines[0].VAT = model.LineVAT{CategoryCode: "O"}
			i.VATBreakdown[0] = model.VATBreakdown{TaxableAmount: "100.00", TaxAmount: "0.00", CategoryCode: "O", ExemptionReasonText: "Not subject to VAT"}
		}},

		// Conditional (BR-CO) rules.
		{"VAT point date and code together", "BR-CO-3", func(i *model.Invoice) { i.TaxPointDate = "2024-01-01"; i.TaxPointDateCode = "5" }},
		{"VAT id without country prefix", "BR-CO-9", func(i *model.Invoice) { i.Seller.VATIdentifier = "123456789" }},
		{"no VAT breakdown", "BR-CO-18", func(i *model.Invoice) { i.VATBreakdown = nil }},
		{"due amount without date or terms", "BR-CO-25", func(i *model.Invoice) { i.DueDate = ""; i.PaymentTerms = "" }},
		{"seller not identified", "BR-CO-26", func(i *model.Invoice) {
			i.Seller.VATIdentifier = ""
			i.Seller.TaxRegistrationID = ""
			i.Seller.Identifiers = nil
			i.Seller.LegalRegistrationID = nil
		}},
		{"invalid item classification scheme", "BR-CL-13", func(i *model.Invoice) {
			i.Lines[0].Item.Classifications = []model.ItemClassification{{Code: "X", ListID: "ZZZZ"}}
		}},

		// Format / structure / further code lists / PEPPOL.
		{"too many decimals on line net", "BR-DEC-23", func(i *model.Invoice) { i.Lines[0].NetAmount = "100.001" }},
		{"negative item net price", "BR-27", func(i *model.Invoice) { i.Lines[0].Price.NetPrice = "-1.00" }},
		{"invoicing period end before start", "BR-29", func(i *model.Invoice) {
			i.Delivery = &model.Delivery{InvoicingPeriod: &model.Period{Start: "2024-03-01", End: "2024-02-01"}}
		}},
		{"item attribute without value", "BR-54", func(i *model.Invoice) {
			i.Lines[0].Item.Attributes = []model.ItemAttribute{{Name: "Farbe", Value: ""}}
		}},
		{"missing seller postal address", "BR-8", func(i *model.Invoice) { i.Seller.Address = model.Address{} }},
		{"invalid VAT point date code", "BR-CL-6", func(i *model.Invoice) { i.TaxPointDate = ""; i.TaxPointDateCode = "ZZ" }},
		{"line net amount mismatch", "PEPPOL-EN16931-R120", func(i *model.Invoice) { i.Lines[0].Quantity.Value = "5" }},
		{"tax currency equals invoice currency", "PEPPOL-EN16931-R005", func(i *model.Invoice) { i.VATAccountingCurrency = "EUR" }},
		{"electronic address without scheme", "BR-62", func(i *model.Invoice) { i.Seller.ElectronicAddress.Scheme = "" }},
		{"direct debit without mandate", "PEPPOL-EN16931-R061", func(i *model.Invoice) {
			i.PaymentInstructions.DirectDebit = &model.DirectDebit{CreditorIdentifier: "DE98ZZZ", DebitedAccount: "DE02120300000000202051"}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inv := validInvoice()
			c.mut(inv)
			if !fires(inv, c.rule) {
				t.Errorf("expected rule %s to fire after %q", c.rule, c.name)
			}
		})
	}
}

func TestResultJSONLocalized(t *testing.T) {
	inv := validInvoice()
	inv.BuyerReference = "" // triggers BR-DE-15 (German national rule)

	de, err := Validate(inv).JSON("de")
	if err != nil {
		t.Fatal(err)
	}
	en, err := Validate(inv).JSON("en")
	if err != nil {
		t.Fatal(err)
	}
	// The German rendering uses the SeMoX German text; English uses the schematron.
	if !strings.Contains(string(de), "übermittelt") {
		t.Errorf("German JSON missing German rule text:\n%s", de)
	}
	if !strings.Contains(string(de), `"valid": false`) {
		t.Errorf("expected valid:false in result")
	}
	if !strings.Contains(string(en), "BR-DE-15") {
		t.Errorf("English JSON missing rule id")
	}
}
