package validation

import (
	"math/big"
	"strconv"

	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(rulesVATCategory)
}

type rateKind int

const (
	ratePositive rateKind = iota // S, L (IGIC), M (IPSI): rate > 0
	rateZero                     // Z, E, AE, K (IC), G: rate = 0
	rateAbsent                   // O: no rate
)

// catBehavior describes how one VAT category code constrains the invoice.
type catBehavior struct {
	family    string   // rule id prefix, e.g. "BR-S"
	rate      rateKind //
	exempt    bool     // a VAT exemption reason (BT-120/121) is required
	forbidEx  bool     // a VAT exemption reason must NOT be present
	sellerVAT bool     // the seller must be VAT-identified
	buyerVAT  bool     // the buyer must be VAT-identified
	noVATIDs  bool     // (O) the seller must NOT be VAT-identified
}

// behaviors maps EN16931 VAT category codes to their rule behavior.
var behaviors = map[string]catBehavior{
	"S":  {family: "BR-S", rate: ratePositive, forbidEx: true, sellerVAT: true},
	"Z":  {family: "BR-Z", rate: rateZero, forbidEx: true, sellerVAT: true},
	"E":  {family: "BR-E", rate: rateZero, exempt: true, sellerVAT: true},
	"AE": {family: "BR-AE", rate: rateZero, exempt: true, sellerVAT: true, buyerVAT: true},
	"K":  {family: "BR-IC", rate: rateZero, exempt: true, sellerVAT: true, buyerVAT: true},
	"G":  {family: "BR-G", rate: rateZero, exempt: true, sellerVAT: true},
	"O":  {family: "BR-O", rate: rateAbsent, exempt: true, noVATIDs: true},
	"L":  {family: "BR-IG", rate: ratePositive, forbidEx: true, sellerVAT: true},
	"M":  {family: "BR-IP", rate: ratePositive, forbidEx: true, sellerVAT: true},
}

func rulesVATCategory(inv *model.Invoice, rp *reporter) {
	inLines, inAllow, inCharge := categoriesBySource(inv)
	used := map[string]bool{}
	for c := range inLines {
		used[c] = true
	}
	for c := range inAllow {
		used[c] = true
	}
	for c := range inCharge {
		used[c] = true
	}
	sellerVAT := sellerVATIdentified(inv)
	buyerVAT := present(inv.Buyer.VATIdentifier)
	taxRepVAT := inv.TaxRepresentative != nil && present(inv.TaxRepresentative.VATIdentifier)

	for code := range used {
		b, ok := behaviors[code]
		if !ok {
			continue
		}
		// BR-X-1: a used category must appear in the VAT breakdown.
		if !breakdownHasCategory(inv, code) {
			rp.fail(b.family+"-1", "BG-23")
		}
		// BR-X-2/3/4: the required identification per source element (line/allowance/
		// charge). For O it is the opposite — a VAT identifier must NOT be present.
		var unmet bool
		switch {
		case b.noVATIDs: // O: any VAT identifier is a violation
			unmet = present(inv.Seller.VATIdentifier) || buyerVAT || taxRepVAT
		case b.buyerVAT: // AE/IC: seller and buyer must be VAT-identified
			unmet = !sellerVAT || !buyerVAT
		case b.sellerVAT: // S/Z/E/G: seller must be VAT-identified
			unmet = !sellerVAT
		}
		if unmet {
			if inLines[code] {
				rp.fail(b.family+"-2", "BT-151")
			}
			if inAllow[code] {
				rp.fail(b.family+"-3", "BT-95")
			}
			if inCharge[code] {
				rp.fail(b.family+"-4", "BT-102")
			}
		}
	}

	// BR-X-5/6/7: per-element VAT rate constraints.
	for i := range inv.Lines {
		c := inv.Lines[i].VAT.CategoryCode
		if b, ok := behaviors[c]; ok && rateViolation(b, inv.Lines[i].VAT.Rate) {
			rp.failf(b.family+"-5", "BT-152", "line "+strconv.Itoa(i+1))
		}
	}
	for i := range inv.Allowances {
		c := inv.Allowances[i].VATCategoryCode
		if b, ok := behaviors[c]; ok && rateViolation(b, inv.Allowances[i].VATRate) {
			rp.failf(b.family+"-6", "BT-96", "allowance "+strconv.Itoa(i+1))
		}
	}
	for i := range inv.Charges {
		c := inv.Charges[i].VATCategoryCode
		if b, ok := behaviors[c]; ok && rateViolation(b, inv.Charges[i].VATRate) {
			rp.failf(b.family+"-7", "BT-103", "charge "+strconv.Itoa(i+1))
		}
	}

	// Breakdown-level rules: exemption reason and the category taxable basis.
	tol := slack(inv.CurrencyCode)
	for i := range inv.VATBreakdown {
		bd := &inv.VATBreakdown[i]
		b, ok := behaviors[bd.CategoryCode]
		if !ok {
			continue
		}
		at := "breakdown " + strconv.Itoa(i+1)
		hasEx := present(bd.ExemptionReasonText) || present(bd.ExemptionReasonCode)
		// BR-X-10: exemption reason present/absent depending on the category.
		if b.exempt && !hasEx {
			rp.failf(b.family+"-10", "BT-120/BT-121", at)
		}
		if b.forbidEx && hasEx {
			rp.failf(b.family+"-10", "BT-120/BT-121", at)
		}
		// BR-X-8: the category taxable amount equals Σ line nets + charges − allowances
		// for the same category and rate.
		expected := categoryBasis(inv, bd.CategoryCode, bd.Rate)
		check(rp, b.family+"-8", "BT-116", expected, rat(bd.TaxableAmount), tol, at)
		// BR-X-9: the category tax amount is taxable × rate (positive categories) or
		// zero (zero-rated / exempt / out-of-scope).
		var expTax *big.Rat
		if b.rate == ratePositive {
			expTax = mul(rat(bd.TaxableAmount), new(big.Rat).Quo(rat(bd.Rate), big.NewRat(100, 1)))
		} else {
			expTax = zero()
		}
		check(rp, b.family+"-9", "BT-117", expTax, rat(bd.TaxAmount), tol, at)
	}

	// BR-IC-11 / BR-IC-12: intra-community supplies need delivery information.
	if breakdownHasCategory(inv, "K") {
		if !hasDeliveryDateOrPeriod(inv) {
			rp.fail("BR-IC-11", "BT-72/BG-14")
		}
		if inv.Delivery == nil || inv.Delivery.Address == nil || !present(inv.Delivery.Address.CountryCode) {
			rp.fail("BR-IC-12", "BT-80")
		}
	}

	// BR-O-11..14: an invoice with an "O" (not subject to VAT) breakdown must not
	// contain any other VAT category in the breakdown, lines, allowances or charges.
	if breakdownHasCategory(inv, "O") {
		for i := range inv.VATBreakdown {
			if inv.VATBreakdown[i].CategoryCode != "O" {
				rp.fail("BR-O-11", "BT-118")
				break
			}
		}
		if anyOtherCategory(inLines, "O") {
			rp.fail("BR-O-12", "BT-151")
		}
		if anyOtherCategory(inAllow, "O") {
			rp.fail("BR-O-13", "BT-95")
		}
		if anyOtherCategory(inCharge, "O") {
			rp.fail("BR-O-14", "BT-102")
		}
	}
}

// anyOtherCategory reports whether the set contains a category other than the
// given one.
func anyOtherCategory(set map[string]bool, only string) bool {
	for c := range set {
		if c != only {
			return true
		}
	}
	return false
}

func rateViolation(b catBehavior, rate model.Percentage) bool {
	switch b.rate {
	case ratePositive:
		return present(string(rate)) && rat(rate).Sign() <= 0
	case rateZero:
		return present(string(rate)) && rat(rate).Sign() != 0
	case rateAbsent:
		return present(string(rate))
	}
	return false
}

// categoryBasis sums line net amounts plus charges minus allowances that share
// the given VAT category code and (numerically equal) rate.
func categoryBasis(inv *model.Invoice, code string, rate model.Percentage) *big.Rat {
	sum := zero()
	for i := range inv.Lines {
		l := &inv.Lines[i]
		if l.VAT.CategoryCode == code && ratesEqual(l.VAT.Rate, rate) {
			sum = add(sum, rat(l.NetAmount))
		}
	}
	for i := range inv.Charges {
		c := &inv.Charges[i]
		if c.VATCategoryCode == code && ratesEqual(c.VATRate, rate) {
			sum = add(sum, rat(c.Amount))
		}
	}
	for i := range inv.Allowances {
		a := &inv.Allowances[i]
		if a.VATCategoryCode == code && ratesEqual(a.VATRate, rate) {
			sum = sub(sum, rat(a.Amount))
		}
	}
	return sum
}

func ratesEqual(a, b model.Percentage) bool { return rat(a).Cmp(rat(b)) == 0 }

func categoriesUsed(inv *model.Invoice) map[string]bool {
	used := map[string]bool{}
	for i := range inv.Lines {
		if c := inv.Lines[i].VAT.CategoryCode; c != "" {
			used[c] = true
		}
	}
	for i := range inv.Allowances {
		if c := inv.Allowances[i].VATCategoryCode; c != "" {
			used[c] = true
		}
	}
	for i := range inv.Charges {
		if c := inv.Charges[i].VATCategoryCode; c != "" {
			used[c] = true
		}
	}
	return used
}

// categoriesBySource returns the set of VAT category codes used in lines,
// document allowances and document charges respectively.
func categoriesBySource(inv *model.Invoice) (lines, allowances, charges map[string]bool) {
	lines, allowances, charges = map[string]bool{}, map[string]bool{}, map[string]bool{}
	for i := range inv.Lines {
		if c := inv.Lines[i].VAT.CategoryCode; c != "" {
			lines[c] = true
		}
	}
	for i := range inv.Allowances {
		if c := inv.Allowances[i].VATCategoryCode; c != "" {
			allowances[c] = true
		}
	}
	for i := range inv.Charges {
		if c := inv.Charges[i].VATCategoryCode; c != "" {
			charges[c] = true
		}
	}
	return lines, allowances, charges
}

func breakdownHasCategory(inv *model.Invoice, code string) bool {
	for i := range inv.VATBreakdown {
		if inv.VATBreakdown[i].CategoryCode == code {
			return true
		}
	}
	return false
}

func sellerVATIdentified(inv *model.Invoice) bool {
	if present(inv.Seller.VATIdentifier) || present(inv.Seller.TaxRegistrationID) {
		return true
	}
	return inv.TaxRepresentative != nil && present(inv.TaxRepresentative.VATIdentifier)
}

func hasDeliveryDateOrPeriod(inv *model.Invoice) bool {
	if d := inv.Delivery; d != nil {
		if present(string(d.ActualDeliveryDate)) || d.InvoicingPeriod != nil {
			return true
		}
	}
	for i := range inv.Lines {
		if inv.Lines[i].Period != nil {
			return true
		}
	}
	return false
}
