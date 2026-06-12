package validation

import (
	"strconv"
	"strings"

	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(rulesFormat)
}

// dec2 flags BR-DEC-* when an amount has more than two fractional digits.
func dec2(rp *reporter, rule, location, detail, value string) {
	if i := strings.IndexByte(value, '.'); i >= 0 && len(value)-i-1 > 2 {
		rp.failf(rule, location, detail)
	}
}

// rulesFormat covers the decimal-places (BR-DEC-*), non-negative, period-order
// and scheme-presence rules.
func rulesFormat(inv *model.Invoice, rp *reporter) {
	t := inv.Totals
	// BR-DEC: document totals (max 2 decimals).
	dec2(rp, "BR-DEC-9", "BT-106", "", string(t.LineNetTotal))
	dec2(rp, "BR-DEC-10", "BT-107", "", string(t.AllowanceTotal))
	dec2(rp, "BR-DEC-11", "BT-108", "", string(t.ChargeTotal))
	dec2(rp, "BR-DEC-12", "BT-109", "", string(t.TaxBasisTotal))
	dec2(rp, "BR-DEC-13", "BT-110", "", string(t.TaxTotal))
	dec2(rp, "BR-DEC-14", "BT-112", "", string(t.GrandTotal))
	dec2(rp, "BR-DEC-15", "BT-111", "", string(t.TaxTotalAccountingCurrency))
	dec2(rp, "BR-DEC-16", "BT-113", "", string(t.PaidAmount))
	dec2(rp, "BR-DEC-17", "BT-114", "", string(t.RoundingAmount))
	dec2(rp, "BR-DEC-18", "BT-115", "", string(t.DuePayableAmount))

	for i := range inv.Allowances {
		a := &inv.Allowances[i]
		at := "allowance " + strconv.Itoa(i+1)
		dec2(rp, "BR-DEC-1", "BT-92", at, string(a.Amount))
		dec2(rp, "BR-DEC-2", "BT-93", at, string(a.BaseAmount))
	}
	for i := range inv.Charges {
		c := &inv.Charges[i]
		at := "charge " + strconv.Itoa(i+1)
		dec2(rp, "BR-DEC-5", "BT-99", at, string(c.Amount))
		dec2(rp, "BR-DEC-6", "BT-100", at, string(c.BaseAmount))
	}
	for i := range inv.VATBreakdown {
		b := &inv.VATBreakdown[i]
		at := "breakdown " + strconv.Itoa(i+1)
		dec2(rp, "BR-DEC-19", "BT-116", at, string(b.TaxableAmount))
		dec2(rp, "BR-DEC-20", "BT-117", at, string(b.TaxAmount))
	}

	for i := range inv.Lines {
		l := &inv.Lines[i]
		at := "line " + strconv.Itoa(i+1)
		dec2(rp, "BR-DEC-23", "BT-131", at, string(l.NetAmount))
		for _, a := range l.Allowances {
			dec2(rp, "BR-DEC-24", "BT-136", at, string(a.Amount))
			dec2(rp, "BR-DEC-25", "BT-137", at, string(a.BaseAmount))
		}
		for _, c := range l.Charges {
			dec2(rp, "BR-DEC-27", "BT-141", at, string(c.Amount))
			dec2(rp, "BR-DEC-28", "BT-142", at, string(c.BaseAmount))
		}

		// BR-27 / BR-28: item net / gross price shall not be negative.
		if rat(l.Price.NetPrice).Sign() < 0 {
			rp.failf("BR-27", "BT-146", at)
		}
		if rat(l.Price.GrossPrice).Sign() < 0 {
			rp.failf("BR-28", "BT-148", at)
		}
		// BR-30: invoice line period end shall be on or after the start.
		if p := l.Period; p != nil && !periodOK(p.Start, p.End) {
			rp.failf("BR-30", "BT-134/BT-135", at)
		}
		// BR-54: each item attribute shall have a name and a value.
		for _, a := range l.Item.Attributes {
			if !present(a.Name) || !present(a.Value) {
				rp.failf("BR-54", "BT-160/BT-161", at)
			}
		}
		// BR-64 / BR-65: item standard / classification identifiers need a scheme.
		if si := l.Item.StandardIdentifier; si != nil && !present(si.Scheme) {
			rp.failf("BR-64", "BT-157-1", at)
		}
		for _, c := range l.Item.Classifications {
			if !present(c.ListID) {
				rp.failf("BR-65", "BT-158-1", at)
			}
		}
	}

	// BR-29: invoicing period end shall be on or after the start.
	if d := inv.Delivery; d != nil && d.InvoicingPeriod != nil && !periodOK(d.InvoicingPeriod.Start, d.InvoicingPeriod.End) {
		rp.fail("BR-29", "BT-73/BT-74")
	}

	// BR-62 / BR-63: seller / buyer electronic address shall have a scheme.
	if ea := inv.Seller.ElectronicAddress; ea != nil && !present(ea.Scheme) {
		rp.fail("BR-62", "BT-34-1")
	}
	if ea := inv.Buyer.ElectronicAddress; ea != nil && !present(ea.Scheme) {
		rp.fail("BR-63", "BT-49-1")
	}

	// Payment instructions (BG-16/17).
	if pi := inv.PaymentInstructions; pi != nil {
		if !present(pi.MeansTypeCode) {
			rp.fail("BR-49", "BT-81")
		}
		for i := range pi.CreditTransfers {
			if !present(pi.CreditTransfers[i].AccountIdentifier) {
				rp.failf("BR-50", "BT-84", "credit transfer "+strconv.Itoa(i+1))
			}
		}
		// BR-61: a credit-transfer payment means requires a payment account (BT-84).
		if creditTransferMeans[pi.MeansTypeCode] && len(pi.CreditTransfers) == 0 {
			rp.fail("BR-61", "BT-84")
		}
	}
}

// creditTransferMeans are the UNCL4461 payment means codes that denote a credit
// transfer (so a payment account identifier is required).
var creditTransferMeans = map[string]bool{"30": true, "58": true, "31": true}

// periodOK reports whether a period's end is on or after its start (or either is
// absent). Dates are ISO "YYYY-MM-DD", so string order is chronological.
func periodOK(start, end model.Date) bool {
	if start == "" || end == "" {
		return true
	}
	return string(start) <= string(end)
}
