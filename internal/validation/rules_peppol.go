package validation

import (
	"math/big"
	"strconv"

	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(rulesPeppolSemantic)
}

// rulesPeppolSemantic covers PEPPOL-EN16931-R* rules that are checkable against
// the semantic model (the syntax-structure ones are out of scope).
func rulesPeppolSemantic(inv *model.Invoice, rp *reporter) {
	tol := slack(inv.CurrencyCode)

	// R055: the VAT total (BT-110) and the accounting-currency VAT total (BT-111)
	// must have the same operational sign.
	if present(string(inv.Totals.TaxTotal)) && present(string(inv.Totals.TaxTotalAccountingCurrency)) {
		if rat(inv.Totals.TaxTotal).Sign()*rat(inv.Totals.TaxTotalAccountingCurrency).Sign() < 0 {
			rp.fail("PEPPOL-EN16931-R055", "BT-110/BT-111")
		}
	}

	// R041/R042: an allowance/charge base amount and percentage must be present
	// together.
	ac := func(loc, detail string, base, pct model.Decimal) {
		if present(string(pct)) && !present(string(base)) {
			rp.failf("PEPPOL-EN16931-R041", loc, detail)
		}
		if present(string(base)) && !present(string(pct)) {
			rp.failf("PEPPOL-EN16931-R042", loc, detail)
		}
	}
	for i := range inv.Allowances {
		a := &inv.Allowances[i]
		ac("BT-93/BT-94", "allowance "+strconv.Itoa(i+1), a.BaseAmount, a.Percentage)
	}
	for i := range inv.Charges {
		c := &inv.Charges[i]
		ac("BT-100/BT-101", "charge "+strconv.Itoa(i+1), c.BaseAmount, c.Percentage)
	}
	for i := range inv.Lines {
		l := &inv.Lines[i]
		at := "line " + strconv.Itoa(i+1)
		for _, a := range l.Allowances {
			ac("BT-137/BT-138", at, a.BaseAmount, a.Percentage)
		}
		for _, c := range l.Charges {
			ac("BT-142/BT-143", at, c.BaseAmount, c.Percentage)
		}

		// R121: the item price base quantity must be greater than zero.
		if bq := l.Price.BaseQuantity; bq != nil && present(string(bq.Value)) && rat(bq.Value).Sign() <= 0 {
			rp.failf("PEPPOL-EN16931-R121", "BT-149", at)
		}

		// R046: item net price equals gross price minus the price discount.
		if present(string(l.Price.GrossPrice)) {
			expected := sub(rat(l.Price.GrossPrice), rat(l.Price.Discount))
			check(rp, "PEPPOL-EN16931-R046", "BT-146", expected, rat(l.Price.NetPrice), tol, at)
		}

		// R130: the price base-quantity unit must equal the invoiced quantity unit.
		if bq := l.Price.BaseQuantity; bq != nil && bq.Unit != "" && bq.Unit != l.Quantity.Unit {
			rp.failf("PEPPOL-EN16931-R130", "BT-150", at)
		}

		// R120: line net amount = quantity × (net price / base quantity)
		//        + Σ line charges − Σ line allowances.
		baseQ := big.NewRat(1, 1)
		if bq := l.Price.BaseQuantity; bq != nil && present(string(bq.Value)) && rat(bq.Value).Sign() != 0 {
			baseQ = rat(bq.Value)
		}
		expected := mul(rat(l.Quantity.Value), new(big.Rat).Quo(rat(l.Price.NetPrice), baseQ))
		for _, ch := range l.Charges {
			expected = add(expected, rat(ch.Amount))
		}
		for _, al := range l.Allowances {
			expected = sub(expected, rat(al.Amount))
		}
		check(rp, "PEPPOL-EN16931-R120", "BT-131", expected, rat(l.NetAmount), tol, at)

		// R110/R111: a line period must lie within the invoicing period (BG-14).
		if ip := invoicingPeriod(inv); ip != nil {
			if p := l.Period; p != nil {
				if present(string(p.Start)) && present(string(ip.Start)) && string(p.Start) < string(ip.Start) {
					rp.failf("PEPPOL-EN16931-R110", "BT-134", at)
				}
				if present(string(p.End)) && present(string(ip.End)) && string(p.End) > string(ip.End) {
					rp.failf("PEPPOL-EN16931-R111", "BT-135", at)
				}
			}
		}
	}
}

func invoicingPeriod(inv *model.Invoice) *model.Period {
	if d := inv.Delivery; d != nil {
		return d.InvoicingPeriod
	}
	return nil
}
