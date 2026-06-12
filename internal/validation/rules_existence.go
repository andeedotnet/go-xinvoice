package validation

import (
	"strconv"

	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(
		rulesDocumentExistence,
		rulesPartyExistence,
		rulesLineExistence,
		rulesVATBreakdownExistence,
		rulesAllowanceChargeExistence,
		rulesReferenceExistence,
	)
}

// rulesDocumentExistence covers the mandatory document-level Business Terms.
func rulesDocumentExistence(inv *model.Invoice, rp *reporter) {
	if !present(inv.SpecificationIdentifier) {
		rp.fail("BR-1", "BT-24")
	}
	if !present(inv.Number) {
		rp.fail("BR-2", "BT-1")
	}
	if !present(string(inv.IssueDate)) {
		rp.fail("BR-3", "BT-2")
	}
	if !present(inv.TypeCode) {
		rp.fail("BR-4", "BT-3")
	}
	if !present(inv.CurrencyCode) {
		rp.fail("BR-5", "BT-5")
	}
	if !present(string(inv.Totals.LineNetTotal)) {
		rp.fail("BR-12", "BT-106")
	}
	if !present(string(inv.Totals.TaxBasisTotal)) {
		rp.fail("BR-13", "BT-109")
	}
	if !present(string(inv.Totals.GrandTotal)) {
		rp.fail("BR-14", "BT-112")
	}
	if !present(string(inv.Totals.DuePayableAmount)) {
		rp.fail("BR-15", "BT-115")
	}
	if len(inv.Lines) == 0 {
		rp.fail("BR-16", "BG-25")
	}
	// BR-53: a tax accounting currency (BT-6) requires the VAT total in that
	// currency (BT-111).
	if present(inv.VATAccountingCurrency) && !present(string(inv.Totals.TaxTotalAccountingCurrency)) {
		rp.fail("BR-53", "BT-111")
	}
}

// rulesPartyExistence covers seller, buyer and the optional payee / tax
// representative parties.
func rulesPartyExistence(inv *model.Invoice, rp *reporter) {
	if !present(inv.Seller.Name) {
		rp.fail("BR-6", "BT-27")
	}
	if !present(inv.Buyer.Name) {
		rp.fail("BR-7", "BT-44")
	}
	if inv.Seller.Address == (model.Address{}) {
		rp.fail("BR-8", "BG-5")
	}
	if !present(inv.Seller.Address.CountryCode) {
		rp.fail("BR-9", "BT-40")
	}
	if inv.Buyer.Address == (model.Address{}) {
		rp.fail("BR-10", "BG-8")
	}
	if !present(inv.Buyer.Address.CountryCode) {
		rp.fail("BR-11", "BT-55")
	}
	if p := inv.Payee; p != nil && !present(p.Name) {
		rp.fail("BR-17", "BT-59")
	}
	if tr := inv.TaxRepresentative; tr != nil {
		if !present(tr.Name) {
			rp.fail("BR-18", "BT-62")
		}
		if tr.Address == (model.Address{}) {
			rp.fail("BR-19", "BG-12")
		}
		if !present(tr.Address.CountryCode) {
			rp.fail("BR-20", "BT-69")
		}
		if !present(tr.VATIdentifier) {
			rp.fail("BR-56", "BT-63")
		}
	}
}

// rulesLineExistence covers the mandatory invoice line Business Terms.
func rulesLineExistence(inv *model.Invoice, rp *reporter) {
	for i := range inv.Lines {
		l := &inv.Lines[i]
		at := "line " + strconv.Itoa(i+1)
		if !present(l.ID) {
			rp.failf("BR-21", "BT-126", at)
		}
		if !present(string(l.Quantity.Value)) {
			rp.failf("BR-22", "BT-129", at)
		}
		if !present(l.Quantity.Unit) {
			rp.failf("BR-23", "BT-130", at)
		}
		if !present(string(l.NetAmount)) {
			rp.failf("BR-24", "BT-131", at)
		}
		if !present(l.Item.Name) {
			rp.failf("BR-25", "BT-153", at)
		}
		if !present(string(l.Price.NetPrice)) {
			rp.failf("BR-26", "BT-146", at)
		}
		if !present(l.VAT.CategoryCode) {
			rp.failf("BR-CO-4", "BT-151", at)
		}
	}
}

// rulesVATBreakdownExistence covers BG-23 VAT breakdown lines.
func rulesVATBreakdownExistence(inv *model.Invoice, rp *reporter) {
	for i := range inv.VATBreakdown {
		b := &inv.VATBreakdown[i]
		at := "breakdown " + strconv.Itoa(i+1)
		if !present(string(b.TaxableAmount)) {
			rp.failf("BR-45", "BT-116", at)
		}
		if !present(string(b.TaxAmount)) {
			rp.failf("BR-46", "BT-117", at)
		}
		if !present(b.CategoryCode) {
			rp.failf("BR-47", "BT-118", at)
		}
		// BR-48: a rate is required for every category except "Not subject to VAT" (O).
		if b.CategoryCode != "O" && !present(string(b.Rate)) {
			rp.failf("BR-48", "BT-119", at)
		}
	}
}

// rulesAllowanceChargeExistence covers document- and line-level allowances and
// charges.
func rulesAllowanceChargeExistence(inv *model.Invoice, rp *reporter) {
	for i := range inv.Allowances {
		a := &inv.Allowances[i]
		at := "allowance " + strconv.Itoa(i+1)
		if !present(string(a.Amount)) {
			rp.failf("BR-31", "BT-92", at)
		}
		if !present(a.VATCategoryCode) {
			rp.failf("BR-32", "BT-95", at)
		}
		if !present(a.Reason) && !present(a.ReasonCode) {
			rp.failf("BR-33", "BT-97/BT-98", at)
			rp.failf("BR-CO-21", "BT-97/BT-98", at)
		}
	}
	for i := range inv.Charges {
		c := &inv.Charges[i]
		at := "charge " + strconv.Itoa(i+1)
		if !present(string(c.Amount)) {
			rp.failf("BR-36", "BT-99", at)
		}
		if !present(c.VATCategoryCode) {
			rp.failf("BR-37", "BT-102", at)
		}
		if !present(c.Reason) && !present(c.ReasonCode) {
			rp.failf("BR-38", "BT-104/BT-105", at)
			rp.failf("BR-CO-22", "BT-104/BT-105", at)
		}
	}
	for i := range inv.Lines {
		l := &inv.Lines[i]
		at := "line " + strconv.Itoa(i+1)
		for _, a := range l.Allowances {
			if !present(string(a.Amount)) {
				rp.failf("BR-41", "BT-136", at)
			}
			if !present(a.Reason) && !present(a.ReasonCode) {
				rp.failf("BR-42", "BT-139/BT-140", at)
				rp.failf("BR-CO-23", "BT-139/BT-140", at)
			}
		}
		for _, c := range l.Charges {
			if !present(string(c.Amount)) {
				rp.failf("BR-43", "BT-141", at)
			}
			if !present(c.Reason) && !present(c.ReasonCode) {
				rp.failf("BR-44", "BT-144/BT-145", at)
				rp.failf("BR-CO-24", "BT-144/BT-145", at)
			}
		}
	}
}

// rulesReferenceExistence covers the mandatory fields of optional reference groups.
func rulesReferenceExistence(inv *model.Invoice, rp *reporter) {
	for i := range inv.PrecedingInvoices {
		if !present(inv.PrecedingInvoices[i].Reference) {
			rp.failf("BR-55", "BT-25", "preceding "+strconv.Itoa(i+1))
		}
	}
	for i := range inv.Documents {
		if !present(inv.Documents[i].Reference) {
			rp.failf("BR-52", "BT-122", "document "+strconv.Itoa(i+1))
		}
	}
	if d := inv.Delivery; d != nil && d.Address != nil {
		if !present(d.Address.CountryCode) {
			rp.fail("BR-57", "BT-80")
		}
	}
}
