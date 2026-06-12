package xinvoice_test

import (
	"fmt"

	xinvoice "github.com/andeedotnet/go-xinvoice"
)

// A small but complete invoice in the syntax-neutral JSON model.
const exampleJSON = `{
  "number": "R-2024-1",
  "issueDate": "2024-01-15",
  "dueDate": "2024-02-15",
  "typeCode": "380",
  "currencyCode": "EUR",
  "buyerReference": "04011000-12345-39",
  "businessProcessType": "urn:fdc:peppol.eu:2017:poacc:billing:01:1.0",
  "specificationIdentifier": "urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0",
  "seller": {"name": "Muster GmbH", "vatIdentifier": "DE123456789",
             "electronicAddress": {"value": "s@muster.de", "scheme": "EM"},
             "address": {"city": "Berlin", "postCode": "10115", "countryCode": "DE"},
             "contact": {"point": "Vertrieb", "phone": "+4930123", "email": "s@muster.de"}},
  "buyer": {"name": "Behörde", "electronicAddress": {"value": "b@amt.de", "scheme": "EM"},
            "address": {"city": "Hamburg", "postCode": "20095", "countryCode": "DE"}},
  "paymentInstructions": {"meansTypeCode": "58", "creditTransfers": [{"accountIdentifier": "DE02120300000000202051"}]},
  "totals": {"lineNetTotal": 100.00, "taxBasisTotal": 100.00, "taxTotal": 19.00,
             "grandTotal": 119.00, "duePayableAmount": 119.00},
  "vatBreakdown": [{"taxableAmount": 100.00, "taxAmount": 19.00, "categoryCode": "S", "rate": 19}],
  "lines": [{"id": "1", "quantity": {"value": 1, "unit": "C62"}, "netAmount": 100.00,
             "price": {"netPrice": 100.00}, "vat": {"categoryCode": "S", "rate": 19},
             "item": {"name": "Widget"}}]
}`

// Convert the JSON model to UBL XML and read it back — proving the model is the
// pivot between syntaxes.
func Example_jsonToUBLandBack() {
	inv, err := xinvoice.FromJSON([]byte(exampleJSON))
	if err != nil {
		panic(err)
	}

	ublXML, err := inv.ToXML(xinvoice.UBL)
	if err != nil {
		panic(err)
	}

	back, err := xinvoice.ParseXML(ublXML) // auto-detects UBL vs CII
	if err != nil {
		panic(err)
	}
	fmt.Println(back.Number, back.Totals.GrandTotal)
	// Output: R-2024-1 119.00
}

// The same model serializes to CII just as well.
func Example_jsonToCII() {
	inv, _ := xinvoice.FromJSON([]byte(exampleJSON))
	ciiXML, _ := inv.ToXML(xinvoice.CII)
	back, _ := xinvoice.ParseXML(ciiXML)
	fmt.Println(back.Lines[0].Item.Name)
	// Output: Widget
}

// Validate an invoice and render findings; here a broken total is detected.
func Example_validate() {
	inv, _ := xinvoice.FromJSON([]byte(exampleJSON))
	fmt.Println("valid:", xinvoice.Validate(inv).Valid())

	inv.Totals.GrandTotal = "200.00" // no longer equals BT-109 + BT-110
	res := xinvoice.Validate(inv)
	fmt.Println("valid:", res.Valid())
	// Output:
	// valid: true
	// valid: false
}

// Inspect findings with the public types and use the checksum helpers.
func Example_findings() {
	inv, _ := xinvoice.FromJSON([]byte(exampleJSON))
	inv.Totals.GrandTotal = "200.00" // break BR-CO-15

	var errors int
	for _, f := range xinvoice.Validate(inv).Findings {
		if f.Severity == xinvoice.SeverityError {
			errors++
		}
	}
	fmt.Println("error findings:", errors > 0)
	fmt.Println("IBAN valid:", xinvoice.ValidIBAN("DE75512108001245126199"))
	// Output:
	// error findings: true
	// IBAN valid: true
}
