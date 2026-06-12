package validation

import (
	"math/big"
	"strconv"

	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(rulesCalculation, rulesAllowanceChargeCalc)
}

// rulesAllowanceChargeCalc checks PEPPOL-EN16931-R040: an allowance/charge amount
// equals base amount × percentage / 100 when both are present.
func rulesAllowanceChargeCalc(inv *model.Invoice, rp *reporter) {
	tol := slack(inv.CurrencyCode)
	chk := func(loc, detail string, amount, base, pct model.Decimal) {
		if !present(string(base)) || !present(string(pct)) {
			return
		}
		expected := mul(rat(base), new(big.Rat).Quo(rat(pct), big.NewRat(100, 1)))
		check(rp, "PEPPOL-EN16931-R040", loc, expected, rat(amount), tol, detail)
	}
	for i := range inv.Allowances {
		a := &inv.Allowances[i]
		chk("BT-92", "allowance "+strconv.Itoa(i+1), a.Amount, a.BaseAmount, a.Percentage)
	}
	for i := range inv.Charges {
		c := &inv.Charges[i]
		chk("BT-99", "charge "+strconv.Itoa(i+1), c.Amount, c.BaseAmount, c.Percentage)
	}
	for i := range inv.Lines {
		at := "line " + strconv.Itoa(i+1)
		for _, a := range inv.Lines[i].Allowances {
			chk("BT-136", at, a.Amount, a.BaseAmount, a.Percentage)
		}
		for _, c := range inv.Lines[i].Charges {
			chk("BT-141", at, c.Amount, c.BaseAmount, c.Percentage)
		}
	}
}

// rulesCalculation checks the document-total and VAT arithmetic (BR-CO-10..17)
// with the official rounding tolerance (slack).
func rulesCalculation(inv *model.Invoice, rp *reporter) {
	tol := slack(inv.CurrencyCode)
	t := inv.Totals

	// BR-CO-10: Sum of line net amounts (BT-106) = Σ line net amount (BT-131).
	if present(string(t.LineNetTotal)) {
		sum := zero()
		for i := range inv.Lines {
			sum = add(sum, rat(inv.Lines[i].NetAmount))
		}
		check(rp, "BR-CO-10", "BT-106", sum, rat(t.LineNetTotal), tol)
	}

	// BR-CO-11/12: allowance/charge totals = Σ of the document allowances/charges.
	allowSum := zero()
	for i := range inv.Allowances {
		allowSum = add(allowSum, rat(inv.Allowances[i].Amount))
	}
	chargeSum := zero()
	for i := range inv.Charges {
		chargeSum = add(chargeSum, rat(inv.Charges[i].Amount))
	}
	if present(string(t.AllowanceTotal)) || len(inv.Allowances) > 0 {
		check(rp, "BR-CO-11", "BT-107", allowSum, rat(t.AllowanceTotal), tol)
	}
	if present(string(t.ChargeTotal)) || len(inv.Charges) > 0 {
		check(rp, "BR-CO-12", "BT-108", chargeSum, rat(t.ChargeTotal), tol)
	}

	// BR-CO-13: BT-109 = BT-106 - BT-107 + BT-108.
	if present(string(t.TaxBasisTotal)) {
		expected := add(sub(rat(t.LineNetTotal), rat(t.AllowanceTotal)), rat(t.ChargeTotal))
		check(rp, "BR-CO-13", "BT-109", expected, rat(t.TaxBasisTotal), tol)
	}

	// BR-CO-14: BT-110 = Σ VAT category tax amount (BT-117).
	if present(string(t.TaxTotal)) {
		sum := zero()
		for i := range inv.VATBreakdown {
			sum = add(sum, rat(inv.VATBreakdown[i].TaxAmount))
		}
		check(rp, "BR-CO-14", "BT-110", sum, rat(t.TaxTotal), tol)
	}

	// BR-CO-15: BT-112 = BT-109 + BT-110.
	if present(string(t.GrandTotal)) {
		expected := add(rat(t.TaxBasisTotal), rat(t.TaxTotal))
		check(rp, "BR-CO-15", "BT-112", expected, rat(t.GrandTotal), tol)
	}

	// BR-CO-16: BT-115 = BT-112 - BT-113 + BT-114.
	if present(string(t.DuePayableAmount)) {
		expected := add(sub(rat(t.GrandTotal), rat(t.PaidAmount)), rat(t.RoundingAmount))
		check(rp, "BR-CO-16", "BT-115", expected, rat(t.DuePayableAmount), tol)
	}

	// BR-CO-17: per VAT breakdown, BT-117 = BT-116 × (BT-119 / 100).
	for i := range inv.VATBreakdown {
		b := &inv.VATBreakdown[i]
		if !present(string(b.TaxAmount)) || !present(string(b.Rate)) {
			continue
		}
		expected := mul(rat(b.TaxableAmount), new(big.Rat).Quo(rat(b.Rate), big.NewRat(100, 1)))
		check(rp, "BR-CO-17", "BT-117", expected, rat(b.TaxAmount), tol, "breakdown "+strconv.Itoa(i+1))
	}
}

// check flags rule when actual is not within tol of expected.
func check(rp *reporter, rule, location string, expected, actual, tol *big.Rat, detail ...string) {
	if within(expected, actual, tol) {
		return
	}
	d := "expected " + f2(expected) + ", got " + f2(actual)
	if len(detail) > 0 && detail[0] != "" {
		d = detail[0] + ": " + d
	}
	rp.failf(rule, location, d)
}
