package validation

import (
	"strconv"

	"github.com/andeedotnet/go-xinvoice/internal/catalog"
	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(rulesConditional)
}

// rulesConditional covers the remaining BR-CO-* conditional rules (the
// calculation ones live in rules_calc.go).
func rulesConditional(inv *model.Invoice, rp *reporter) {
	// BR-CO-3: VAT point date (BT-7) and VAT point date code (BT-8) are mutually
	// exclusive.
	if present(string(inv.TaxPointDate)) && present(inv.TaxPointDateCode) {
		rp.fail("BR-CO-3", "BT-7/BT-8")
	}

	// BR-CO-9: each VAT identifier shall start with an ISO 3166-1 country prefix.
	checkVATPrefix(rp, "BT-31", inv.Seller.VATIdentifier)
	checkVATPrefix(rp, "BT-48", inv.Buyer.VATIdentifier)
	if tr := inv.TaxRepresentative; tr != nil {
		checkVATPrefix(rp, "BT-63", tr.VATIdentifier)
	}

	// BR-CO-18: an invoice shall have at least one VAT breakdown (BG-23).
	if len(inv.VATBreakdown) == 0 {
		rp.fail("BR-CO-18", "BG-23")
	}

	// BR-CO-19: an invoicing period (BG-14) shall have a start or end date.
	if d := inv.Delivery; d != nil && d.InvoicingPeriod != nil {
		p := d.InvoicingPeriod
		if !present(string(p.Start)) && !present(string(p.End)) {
			rp.fail("BR-CO-19", "BT-73/BT-74")
		}
	}
	// BR-CO-20: an invoice line period (BG-26) shall have a start or end date.
	for i := range inv.Lines {
		if p := inv.Lines[i].Period; p != nil && !present(string(p.Start)) && !present(string(p.End)) {
			rp.failf("BR-CO-20", "BT-134/BT-135", "line "+strconv.Itoa(i+1))
		}
	}

	// BR-CO-25: if the amount due for payment (BT-115) is positive, a payment due
	// date (BT-9) or payment terms (BT-20) shall be present.
	if due := rat(inv.Totals.DuePayableAmount); due.Sign() > 0 {
		if !present(string(inv.DueDate)) && !present(inv.PaymentTerms) {
			rp.fail("BR-CO-25", "BT-9/BT-20")
		}
	}

	// BR-CO-26: the seller shall be identified by a seller identifier (BT-29), a
	// legal registration identifier (BT-30) or a VAT identifier (BT-31).
	if len(inv.Seller.Identifiers) == 0 && inv.Seller.LegalRegistrationID == nil && !present(inv.Seller.VATIdentifier) {
		rp.fail("BR-CO-26", "BT-29/BT-30/BT-31")
	}
}

// checkVATPrefix flags BR-CO-9 when a VAT identifier does not begin with a valid
// country prefix (ISO 3166-1 alpha-2, plus "EL" for Greece).
func checkVATPrefix(rp *reporter, location, vat string) {
	if !present(vat) {
		return
	}
	if len(vat) < 2 {
		rp.fail("BR-CO-9", location)
		return
	}
	prefix := vat[:2]
	if prefix == "EL" {
		return
	}
	if in, known := catalog.InCodeList("BR-CL-14", prefix); known && !in {
		rp.fail("BR-CO-9", location)
	}
}
