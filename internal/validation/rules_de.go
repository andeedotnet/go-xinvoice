package validation

import (
	"strings"

	"github.com/andeedotnet/go-xinvoice/internal/checksum"
	"github.com/andeedotnet/go-xinvoice/model"
)

func init() {
	register(rulesXRechnungDE, rulesPeppol)
}

// rulesXRechnungDE covers a subset of the German national (BR-DE-*) rules.
func rulesXRechnungDE(inv *model.Invoice, rp *reporter) {
	// BR-DE-1: an invoice must contain PAYMENT INSTRUCTIONS (BG-16).
	if inv.PaymentInstructions == nil {
		rp.fail("BR-DE-1", "BG-16")
	}
	// BR-DE-15: the buyer reference (BT-10) must be provided.
	if !present(inv.BuyerReference) {
		rp.fail("BR-DE-15", "BT-10")
	}

	// Seller postal address detail (BR-DE-3/4) and contact (BR-DE-2/5/6/7).
	if !present(inv.Seller.Address.City) {
		rp.fail("BR-DE-3", "BT-37")
	}
	if !present(inv.Seller.Address.PostCode) {
		rp.fail("BR-DE-4", "BT-38")
	}
	if inv.Seller.Contact == nil {
		rp.fail("BR-DE-2", "BG-6")
	} else {
		if !present(inv.Seller.Contact.Point) {
			rp.fail("BR-DE-5", "BT-41")
		}
		if !present(inv.Seller.Contact.Phone) {
			rp.fail("BR-DE-6", "BT-42")
		}
		if !present(inv.Seller.Contact.Email) {
			rp.fail("BR-DE-7", "BT-43")
		}
	}

	// Buyer postal address detail (BR-DE-8/9).
	if !present(inv.Buyer.Address.City) {
		rp.fail("BR-DE-8", "BT-52")
	}
	if !present(inv.Buyer.Address.PostCode) {
		rp.fail("BR-DE-9", "BT-53")
	}

	// BR-DE-16: when a taxable VAT category is used, the seller must be
	// identified for VAT (BT-31 or BT-32) or have a tax representative (BG-11).
	if usesTaxableCategory(inv) {
		if !present(inv.Seller.VATIdentifier) && !present(inv.Seller.TaxRegistrationID) && inv.TaxRepresentative == nil {
			rp.fail("BR-DE-16", "BT-31/BT-32/BG-11")
		}
	}

	// Deliver-to address detail (BR-DE-10/11) when a deliver-to address (BG-15) is present.
	if d := inv.Delivery; d != nil && d.Address != nil {
		if !present(d.Address.City) {
			rp.fail("BR-DE-10", "BT-77")
		}
		if !present(d.Address.PostCode) {
			rp.fail("BR-DE-11", "BT-78")
		}
	}

	// Direct debit (BG-19): creditor (BT-90) and debited account (BT-91) required.
	if dd := directDebit(inv); dd != nil {
		if !present(dd.CreditorIdentifier) {
			rp.fail("BR-DE-30", "BT-90")
		}
		if !present(dd.DebitedAccount) {
			rp.fail("BR-DE-31", "BT-91")
		}
	}

	// Warnings.
	if !invoiceTypeCodesDE[inv.TypeCode] {
		rp.fail("BR-DE-17", "BT-3") // recommended UNTDID 1001 subset
	}
	if !strings.Contains(strings.ToLower(inv.SpecificationIdentifier), "xrechnung") {
		rp.fail("BR-DE-21", "BT-24")
	}
	if inv.TypeCode == "384" && len(inv.PrecedingInvoices) == 0 {
		rp.fail("BR-DE-26", "BG-3")
	}

	// BR-DE-22: attachment filenames (BT-125-2) must be unique.
	seen := map[string]bool{}
	for i := range inv.Documents {
		if a := inv.Documents[i].Attachment; a != nil && a.Filename != "" {
			if seen[a.Filename] {
				rp.fail("BR-DE-22", "BT-125-2")
			}
			seen[a.Filename] = true
		}
	}
	// BR-DE-27 / BR-DE-28 (warnings): seller contact phone / email plausibility.
	if c := inv.Seller.Contact; c != nil {
		if present(c.Phone) && countDigits(c.Phone) < 3 {
			rp.fail("BR-DE-27", "BT-42")
		}
		if present(c.Email) && !validEmail(c.Email) {
			rp.fail("BR-DE-28", "BT-43")
		}
	}
	if pi := inv.PaymentInstructions; pi != nil {
		if pi.MeansTypeCode == "58" {
			for _, ct := range pi.CreditTransfers {
				if present(ct.AccountIdentifier) && !checksum.ValidIBAN(ct.AccountIdentifier) {
					rp.fail("BR-DE-19", "BT-84")
				}
			}
		}
		if pi.MeansTypeCode == "59" && pi.DirectDebit != nil && present(pi.DirectDebit.DebitedAccount) && !checksum.ValidIBAN(pi.DirectDebit.DebitedAccount) {
			rp.fail("BR-DE-20", "BT-91")
		}
	}
}

// invoiceTypeCodesDE is the UNTDID 1001 subset recommended by XRechnung (BR-DE-17).
var invoiceTypeCodesDE = map[string]bool{
	"326": true, "380": true, "384": true, "389": true, "381": true,
	"875": true, "876": true, "877": true,
}

func countDigits(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n++
		}
	}
	return n
}

// validEmail does a minimal plausibility check (one "@" with non-empty parts and
// a dotted domain).
func validEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	return strings.Contains(s[at+1:], ".")
}

// directDebit returns the direct-debit information (BG-19) if present.
func directDebit(inv *model.Invoice) *model.DirectDebit {
	if pi := inv.PaymentInstructions; pi != nil {
		return pi.DirectDebit
	}
	return nil
}

// rulesPeppol covers a subset of the PEPPOL-EN16931-R* rules.
func rulesPeppol(inv *model.Invoice, rp *reporter) {
	// PEPPOL-EN16931-R001: business process (BT-23) must be provided.
	if !present(inv.BusinessProcessType) {
		rp.fail("PEPPOL-EN16931-R001", "BT-23")
	}
	// PEPPOL-EN16931-R020 / R010: seller / buyer electronic address must be provided.
	if inv.Seller.ElectronicAddress == nil || !present(inv.Seller.ElectronicAddress.Value) {
		rp.fail("PEPPOL-EN16931-R020", "BT-34")
	}
	if inv.Buyer.ElectronicAddress == nil || !present(inv.Buyer.ElectronicAddress.Value) {
		rp.fail("PEPPOL-EN16931-R010", "BT-49")
	}
	// PEPPOL-EN16931-R005: the VAT accounting currency (BT-6) must differ from the
	// invoice currency (BT-5) when provided.
	if present(inv.VATAccountingCurrency) && inv.VATAccountingCurrency == inv.CurrencyCode {
		rp.fail("PEPPOL-EN16931-R005", "BT-6")
	}
	// PEPPOL-EN16931-R061: a mandate reference (BT-89) is required for direct debit.
	if dd := directDebit(inv); dd != nil && !present(dd.MandateReference) {
		rp.fail("PEPPOL-EN16931-R061", "BT-89")
	}
}

// usesTaxableCategory reports whether any taxable VAT category code (one that
// requires seller VAT identification) appears in the invoice.
func usesTaxableCategory(inv *model.Invoice) bool {
	taxable := map[string]bool{"S": true, "Z": true, "E": true, "AE": true, "K": true, "G": true, "L": true, "M": true}
	for i := range inv.VATBreakdown {
		if taxable[inv.VATBreakdown[i].CategoryCode] {
			return true
		}
	}
	for i := range inv.Lines {
		if taxable[inv.Lines[i].VAT.CategoryCode] {
			return true
		}
	}
	for i := range inv.Allowances {
		if taxable[inv.Allowances[i].VATCategoryCode] {
			return true
		}
	}
	for i := range inv.Charges {
		if taxable[inv.Charges[i].VATCategoryCode] {
			return true
		}
	}
	return false
}
