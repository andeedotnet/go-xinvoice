package validation

import (
	"strconv"

	"github.com/andeedotnet/go-xinvoice/internal/catalog"
	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(rulesCodeList, rulesCodeListMore)
}

// rulesCodeListMore covers the remaining BR-CL-* scheme/code checks.
func rulesCodeListMore(inv *model.Invoice, rp *reporter) {
	checkCode(rp, "BR-CL-3", "BT-5", "", inv.CurrencyCode)     // amount currencyID
	checkCode(rp, "BR-CL-6", "BT-8", "", inv.TaxPointDateCode) // UNTDID 2005
	if o := inv.InvoicedObject; o != nil {
		checkCode(rp, "BR-CL-7", "BT-18-1", "", o.Scheme) // UNTDID 1153
	}
	for i := range inv.Notes {
		checkCode(rp, "BR-CL-8", "BT-21", "note "+strconv.Itoa(i+1), inv.Notes[i].SubjectCode) // UNCL4451
	}
	// Party identifier schemes (ISO 6523 ICD).
	for _, id := range inv.Seller.Identifiers {
		checkCode(rp, "BR-CL-10", "BT-29-1", "seller", id.Scheme)
	}
	for _, id := range inv.Buyer.Identifiers {
		checkCode(rp, "BR-CL-10", "BT-46-1", "buyer", id.Scheme)
	}
	// Registration identifier schemes (ISO 6523 ICD).
	if r := inv.Seller.LegalRegistrationID; r != nil {
		checkCode(rp, "BR-CL-11", "BT-30-1", "seller", r.Scheme)
	}
	if r := inv.Buyer.LegalRegistrationID; r != nil {
		checkCode(rp, "BR-CL-11", "BT-47-1", "buyer", r.Scheme)
	}
	// Delivery location identifier scheme (ISO 6523 ICD).
	if d := inv.Delivery; d != nil && d.LocationIdentifier != nil {
		checkCode(rp, "BR-CL-26", "BT-71-1", "", d.LocationIdentifier.Scheme)
	}
	// Line VAT category (BR-CL-18 duplicates BR-CL-17 at line level).
	for i := range inv.Lines {
		checkCode(rp, "BR-CL-18", "BT-151", "line "+strconv.Itoa(i+1), inv.Lines[i].VAT.CategoryCode)
	}
}

// checkCode flags a finding when value is present, the rule's code list is known,
// and value is not a member of it.
func checkCode(rp *reporter, rule, location, detail, value string) {
	if !present(value) {
		return
	}
	if in, known := catalog.InCodeList(rule, value); known && !in {
		rp.failf(rule, location, detail)
	}
}

// rulesCodeList enforces the BR-CL-* code-list memberships using the catalog.
func rulesCodeList(inv *model.Invoice, rp *reporter) {
	checkCode(rp, "BR-CL-01", "BT-3", "", inv.TypeCode)
	checkCode(rp, "BR-CL-04", "BT-5", "", inv.CurrencyCode)
	checkCode(rp, "BR-CL-05", "BT-6", "", inv.VATAccountingCurrency)

	// Country codes (ISO 3166).
	checkCode(rp, "BR-CL-14", "BT-40", "seller", inv.Seller.Address.CountryCode)
	checkCode(rp, "BR-CL-14", "BT-55", "buyer", inv.Buyer.Address.CountryCode)
	if tr := inv.TaxRepresentative; tr != nil {
		checkCode(rp, "BR-CL-14", "BT-69", "tax representative", tr.Address.CountryCode)
	}
	if d := inv.Delivery; d != nil && d.Address != nil {
		checkCode(rp, "BR-CL-15", "BT-80", "deliver to", d.Address.CountryCode)
	}

	// Payment means (UNCL4461).
	if pi := inv.PaymentInstructions; pi != nil {
		checkCode(rp, "BR-CL-16", "BT-81", "", pi.MeansTypeCode)
	}

	// Electronic address schemes (EAS).
	if ea := inv.Seller.ElectronicAddress; ea != nil {
		checkCode(rp, "BR-CL-25", "BT-34-1", "seller", ea.Scheme)
	}
	if ea := inv.Buyer.ElectronicAddress; ea != nil {
		checkCode(rp, "BR-CL-25", "BT-49-1", "buyer", ea.Scheme)
	}

	// VAT category codes (UNCL5305).
	for i := range inv.Allowances {
		checkCode(rp, "BR-CL-17", "BT-95", "allowance "+strconv.Itoa(i+1), inv.Allowances[i].VATCategoryCode)
	}
	for i := range inv.Charges {
		checkCode(rp, "BR-CL-17", "BT-102", "charge "+strconv.Itoa(i+1), inv.Charges[i].VATCategoryCode)
	}
	for i := range inv.VATBreakdown {
		checkCode(rp, "BR-CL-17", "BT-118", "breakdown "+strconv.Itoa(i+1), inv.VATBreakdown[i].CategoryCode)
		checkCode(rp, "BR-CL-22", "BT-121", "breakdown "+strconv.Itoa(i+1), inv.VATBreakdown[i].ExemptionReasonCode)
	}

	// Allowance / charge reason codes (UNCL5189 / UNCL7161).
	for i := range inv.Allowances {
		checkCode(rp, "BR-CL-19", "BT-98", "allowance "+strconv.Itoa(i+1), inv.Allowances[i].ReasonCode)
	}
	for i := range inv.Charges {
		checkCode(rp, "BR-CL-20", "BT-105", "charge "+strconv.Itoa(i+1), inv.Charges[i].ReasonCode)
	}

	// Line-level codes.
	for i := range inv.Lines {
		l := &inv.Lines[i]
		at := "line " + strconv.Itoa(i+1)
		checkCode(rp, "BR-CL-17", "BT-151", at, l.VAT.CategoryCode)
		checkCode(rp, "BR-CL-23", "BT-130", at, l.Quantity.Unit)
		if l.Price.BaseQuantity != nil {
			checkCode(rp, "BR-CL-23", "BT-150", at, l.Price.BaseQuantity.Unit)
		}
		if l.Item.StandardIdentifier != nil {
			checkCode(rp, "BR-CL-21", "BT-157-1", at, l.Item.StandardIdentifier.Scheme)
		}
		for _, c := range l.Item.Classifications {
			checkCode(rp, "BR-CL-13", "BT-158-1", at, c.ListID)
		}
		checkCode(rp, "BR-CL-14", "BT-159", at, l.Item.CountryOfOrigin)
	}
}
