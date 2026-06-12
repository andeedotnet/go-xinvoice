package ubl

import (
	"regexp"

	"github.com/andeedotnet/go-xinvoice/model"
)

// notePrefixRe matches the EN16931 UBL convention of encoding the note subject
// code (BT-21) as a "#CODE#" prefix on the note text (BT-22).
var notePrefixRe = regexp.MustCompile(`(?s)^#([0-9A-Za-z]+)#(.*)$`)

// toNote splits a "#CODE#text" UBL note into subject code and text.
func toNote(s string) model.Note {
	if m := notePrefixRe.FindStringSubmatch(s); m != nil {
		return model.Note{SubjectCode: m[1], Text: m[2]}
	}
	return model.Note{Text: s}
}

// toModel converts a parsed UBL [Invoice] into the syntax-neutral model.
func toModel(u *Invoice) *model.Invoice {
	inv := &model.Invoice{
		SpecificationIdentifier:  u.CustomizationID,
		BusinessProcessType:      u.ProfileID,
		Number:                   u.ID,
		IssueDate:                model.Date(u.IssueDate),
		DueDate:                  model.Date(u.DueDate),
		TypeCode:                 firstNonEmpty(u.InvoiceTypeCode, u.CreditNoteTypeCode),
		TaxPointDate:             model.Date(u.TaxPointDate),
		CurrencyCode:             u.DocumentCurrencyCode,
		VATAccountingCurrency:    u.TaxCurrencyCode,
		BuyerAccountingReference: u.AccountingCost,
		BuyerReference:           u.BuyerReference,
		Seller:                   toParty(&u.AccountingSupplierParty.Party),
		Buyer:                    toParty(&u.AccountingCustomerParty.Party),
	}

	for _, n := range u.Note {
		inv.Notes = append(inv.Notes, toNote(n))
	}
	if u.OrderReference != nil {
		inv.PurchaseOrderReference = u.OrderReference.ID
		inv.SalesOrderReference = u.OrderReference.SalesOrderID
	}
	if u.ContractDocumentReference != nil {
		inv.ContractReference = u.ContractDocumentReference.ID
	}
	if u.ProjectReference != nil {
		inv.ProjectReference = u.ProjectReference.ID
	}
	if u.DespatchDocumentReference != nil {
		inv.DespatchAdviceReference = u.DespatchDocumentReference.ID
	}
	if u.ReceiptDocumentReference != nil {
		inv.ReceivingAdviceReference = u.ReceiptDocumentReference.ID
	}
	if u.OriginatorDocumentReference != nil {
		inv.TenderOrLotReference = u.OriginatorDocumentReference.ID
	}
	for _, br := range u.BillingReference {
		inv.PrecedingInvoices = append(inv.PrecedingInvoices, model.PrecedingInvoiceReference{
			Reference: br.InvoiceDocumentReference.ID,
			IssueDate: model.Date(br.InvoiceDocumentReference.IssueDate),
		})
	}
	for _, d := range u.AdditionalDocumentReference {
		if d.DocumentTypeCode == "130" {
			inv.InvoicedObject = &model.ID{Value: d.ID.Value, Scheme: d.ID.SchemeID}
			continue
		}
		sd := model.SupportingDocument{
			Reference:   d.ID.Value,
			Description: d.DocumentDescription,
		}
		if d.Attachment != nil {
			if d.Attachment.ExternalReference != nil {
				sd.ExternalLocation = d.Attachment.ExternalReference.URI
			}
			if b := d.Attachment.EmbeddedDocumentBinaryObject; b != nil {
				sd.Attachment = &model.BinaryObject{
					MimeCode: b.MimeCode,
					Filename: b.Filename,
					Content:  []byte(b.Value),
				}
			}
		}
		inv.Documents = append(inv.Documents, sd)
	}
	for _, pp := range u.PrepaidPayment {
		inv.ThirdPartyPayments = append(inv.ThirdPartyPayments, model.ThirdPartyPayment{
			Type:        pp.ID,
			Amount:      model.Amount(pp.PaidAmount.Value),
			Description: pp.InstructionID,
		})
	}

	if u.PayeeParty != nil {
		inv.Payee = toPayee(u.PayeeParty)
	}
	if u.TaxRepresentativeParty != nil {
		inv.TaxRepresentative = toTaxRep(u.TaxRepresentativeParty)
	}
	if u.Delivery != nil {
		inv.Delivery = toDelivery(u.Delivery, u.InvoicePeriod)
	} else if u.InvoicePeriod != nil {
		inv.Delivery = &model.Delivery{InvoicingPeriod: toPeriod(u.InvoicePeriod)}
	}
	inv.PaymentInstructions = toPaymentInstructions(u.PaymentMeans)
	if u.PaymentTerms != nil {
		inv.PaymentTerms = u.PaymentTerms.Note
	}

	for _, ac := range u.AllowanceCharge {
		m := toAllowanceCharge(ac)
		if ac.ChargeIndicator {
			inv.Charges = append(inv.Charges, m)
		} else {
			inv.Allowances = append(inv.Allowances, m)
		}
	}

	inv.Totals = toTotals(u.LegalMonetaryTotal)
	inv.VATBreakdown, inv.Totals.TaxTotal, inv.Totals.TaxTotalAccountingCurrency = toTax(u.TaxTotal)

	lines := u.InvoiceLine
	if len(lines) == 0 {
		lines = u.CreditNoteLine
	}
	for _, l := range lines {
		inv.Lines = append(inv.Lines, toLine(l))
	}
	extractSEPACreditor(inv)
	return inv
}

// extractSEPACreditor moves a SEPA-schemed party identifier (BT-90, bank assigned
// creditor identifier) from the seller/payee party into the direct-debit
// information, matching the syntax-neutral model (and the CII binding).
func extractSEPACreditor(inv *model.Invoice) {
	creditor := ""
	kept := inv.Seller.Identifiers[:0]
	for _, id := range inv.Seller.Identifiers {
		if id.Scheme == "SEPA" {
			creditor = id.Value
		} else {
			kept = append(kept, id)
		}
	}
	inv.Seller.Identifiers = kept
	if p := inv.Payee; p != nil && p.Identifier != nil && p.Identifier.Scheme == "SEPA" {
		creditor = p.Identifier.Value
		p.Identifier = nil
	}
	if creditor == "" {
		return
	}
	if inv.PaymentInstructions == nil {
		inv.PaymentInstructions = &model.PaymentInstructions{}
	}
	if inv.PaymentInstructions.DirectDebit == nil {
		inv.PaymentInstructions.DirectDebit = &model.DirectDebit{}
	}
	inv.PaymentInstructions.DirectDebit.CreditorIdentifier = creditor
}

// toQuantity converts a UBL quantity (nil-safe) to the model quantity.
func toQuantity(q *quantity) model.Quantity {
	if q == nil {
		return model.Quantity{}
	}
	return model.Quantity{Value: model.Decimal(q.Value), Unit: q.Unit}
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func toParty(p *Party) model.Party {
	mp := model.Party{
		Address: toAddress(p.PostalAddress),
		Contact: toContact(p.Contact),
	}
	if p.PartyName != nil {
		mp.TradingName = p.PartyName.Name
	}
	if p.EndpointID != nil {
		mp.ElectronicAddress = &model.ID{Value: p.EndpointID.Value, Scheme: p.EndpointID.SchemeID}
	}
	for _, pid := range p.PartyIdentification {
		mp.Identifiers = append(mp.Identifiers, model.ID{Value: pid.ID.Value, Scheme: pid.ID.SchemeID})
	}
	if le := p.PartyLegalEntity; le != nil {
		mp.Name = le.RegistrationName
		mp.AdditionalLegalInfo = le.CompanyLegalForm
		if le.CompanyID != nil {
			mp.LegalRegistrationID = &model.ID{Value: le.CompanyID.Value, Scheme: le.CompanyID.SchemeID}
		}
	}
	for _, ts := range p.PartyTaxScheme {
		if ts.TaxScheme.ID == "VAT" {
			mp.VATIdentifier = ts.CompanyID
		} else {
			mp.TaxRegistrationID = ts.CompanyID
		}
	}
	return mp
}

func toPayee(p *Party) *model.Payee {
	pe := &model.Payee{}
	if p.PartyName != nil {
		pe.Name = p.PartyName.Name
	}
	if len(p.PartyIdentification) > 0 {
		id := p.PartyIdentification[0].ID
		pe.Identifier = &model.ID{Value: id.Value, Scheme: id.SchemeID}
	}
	if le := p.PartyLegalEntity; le != nil && le.CompanyID != nil {
		pe.LegalRegistrationID = &model.ID{Value: le.CompanyID.Value, Scheme: le.CompanyID.SchemeID}
	}
	return pe
}

func toTaxRep(p *Party) *model.TaxRepresentative {
	tr := &model.TaxRepresentative{Address: toAddress(p.PostalAddress)}
	if p.PartyName != nil {
		tr.Name = p.PartyName.Name
	}
	for _, ts := range p.PartyTaxScheme {
		if ts.TaxScheme.ID == "VAT" {
			tr.VATIdentifier = ts.CompanyID
		}
	}
	return tr
}

func toAddress(a *Address) model.Address {
	if a == nil {
		return model.Address{}
	}
	ma := model.Address{
		Line1:       a.StreetName,
		Line2:       a.AdditionalStreetName,
		City:        a.CityName,
		PostCode:    a.PostalZone,
		Subdivision: a.CountrySubentity,
		CountryCode: a.Country.IdentificationCode,
	}
	if a.AddressLine != nil {
		ma.Line3 = a.AddressLine.Line
	}
	return ma
}

func toContact(c *Contact) *model.Contact {
	if c == nil {
		return nil
	}
	return &model.Contact{Point: c.Name, Phone: c.Telephone, Email: c.ElectronicMail}
}

func toDelivery(d *Delivery, period *Period) *model.Delivery {
	md := &model.Delivery{
		ActualDeliveryDate: model.Date(d.ActualDeliveryDate),
		InvoicingPeriod:    toPeriod(period),
	}
	if loc := d.DeliveryLocation; loc != nil {
		if loc.ID != nil {
			md.LocationIdentifier = &model.ID{Value: loc.ID.Value, Scheme: loc.ID.SchemeID}
		}
		if loc.Address != nil {
			a := toAddress(loc.Address)
			md.Address = &a
		}
	}
	if dp := d.DeliveryParty; dp != nil && dp.PartyName != nil {
		md.PartyName = dp.PartyName.Name
	}
	return md
}

func toPeriod(p *Period) *model.Period {
	if p == nil || (p.StartDate == "" && p.EndDate == "") {
		return nil
	}
	return &model.Period{Start: model.Date(p.StartDate), End: model.Date(p.EndDate)}
}

func toPaymentInstructions(means []PaymentMeans) *model.PaymentInstructions {
	if len(means) == 0 {
		return nil
	}
	pi := &model.PaymentInstructions{
		MeansTypeCode: means[0].PaymentMeansCode.Value,
		MeansText:     means[0].PaymentMeansCode.Name,
	}
	for _, m := range means {
		if pi.RemittanceInformation == "" {
			pi.RemittanceInformation = m.PaymentID
		}
		if fa := m.PayeeFinancialAccount; fa != nil {
			ct := model.CreditTransfer{AccountIdentifier: fa.ID, AccountName: fa.Name}
			if fa.FinancialInstitutionBranch != nil {
				ct.ServiceProviderID = fa.FinancialInstitutionBranch.ID
			}
			pi.CreditTransfers = append(pi.CreditTransfers, ct)
		}
		if ca := m.CardAccount; ca != nil && pi.Card == nil {
			pi.Card = &model.PaymentCard{PrimaryAccountNumber: ca.PrimaryAccountNumberID, HolderName: ca.HolderName}
		}
		if pm := m.PaymentMandate; pm != nil && pi.DirectDebit == nil {
			dd := &model.DirectDebit{MandateReference: pm.ID}
			if pm.PayerFinancialAccount != nil {
				dd.DebitedAccount = pm.PayerFinancialAccount.ID
			}
			pi.DirectDebit = dd
		}
	}
	return pi
}

func toAllowanceCharge(ac AllowanceCharge) model.AllowanceCharge {
	m := model.AllowanceCharge{
		Amount:     model.Amount(ac.Amount.Value),
		Percentage: model.Percentage(ac.MultiplierFactorNumeric),
		Reason:     ac.AllowanceChargeReason,
		ReasonCode: ac.AllowanceChargeReasonCode,
	}
	if ac.BaseAmount != nil {
		m.BaseAmount = model.Amount(ac.BaseAmount.Value)
	}
	if ac.TaxCategory != nil {
		m.VATCategoryCode = ac.TaxCategory.ID
		m.VATRate = model.Percentage(ac.TaxCategory.Percent)
	}
	return m
}

func toTotals(t MonetaryTotal) model.DocumentTotals {
	dt := model.DocumentTotals{
		LineNetTotal:     model.Amount(t.LineExtensionAmount.Value),
		TaxBasisTotal:    model.Amount(t.TaxExclusiveAmount.Value),
		GrandTotal:       model.Amount(t.TaxInclusiveAmount.Value),
		DuePayableAmount: model.Amount(t.PayableAmount.Value),
	}
	if t.AllowanceTotalAmount != nil {
		dt.AllowanceTotal = model.Amount(t.AllowanceTotalAmount.Value)
	}
	if t.ChargeTotalAmount != nil {
		dt.ChargeTotal = model.Amount(t.ChargeTotalAmount.Value)
	}
	if t.PrepaidAmount != nil {
		dt.PaidAmount = model.Amount(t.PrepaidAmount.Value)
	}
	if t.PayableRoundingAmount != nil {
		dt.RoundingAmount = model.Amount(t.PayableRoundingAmount.Value)
	}
	return dt
}

// toTax maps the UBL tax totals to the VAT breakdown plus the document-level VAT
// amount (BT-110) and the accounting-currency VAT amount (BT-111). A TaxTotal
// without subtotals carries BT-111.
func toTax(totals []TaxTotal) (breakdown []model.VATBreakdown, taxTotal, taxTotalAcct model.Amount) {
	for _, tt := range totals {
		if len(tt.TaxSubtotal) == 0 {
			taxTotalAcct = model.Amount(tt.TaxAmount.Value)
			continue
		}
		taxTotal = model.Amount(tt.TaxAmount.Value)
		for _, st := range tt.TaxSubtotal {
			breakdown = append(breakdown, model.VATBreakdown{
				TaxableAmount:       model.Amount(st.TaxableAmount.Value),
				TaxAmount:           model.Amount(st.TaxAmount.Value),
				CategoryCode:        st.TaxCategory.ID,
				Rate:                model.Percentage(st.TaxCategory.Percent),
				ExemptionReasonText: st.TaxCategory.TaxExemptionReason,
				ExemptionReasonCode: st.TaxCategory.TaxExemptionReasonCode,
			})
		}
	}
	return breakdown, taxTotal, taxTotalAcct
}

func toLine(l InvoiceLine) model.Line {
	q := l.InvoicedQuantity
	if q == nil {
		q = l.CreditedQuantity
	}
	ml := model.Line{
		ID:                       l.ID,
		Note:                     l.Note,
		Quantity:                 toQuantity(q),
		NetAmount:                model.Amount(l.LineExtensionAmount.Value),
		BuyerAccountingReference: l.AccountingCost,
		Period:                   toPeriod(l.InvoicePeriod),
		Item:                     toItem(l.Item),
		Price:                    toPrice(l.Price),
	}
	if l.OrderLineReference != nil {
		ml.OrderLineReference = l.OrderLineReference.LineID
	}
	if l.DocumentReference != nil {
		ml.ObjectIdentifier = &model.ID{Value: l.DocumentReference.ID.Value, Scheme: l.DocumentReference.ID.SchemeID}
	}
	ml.VAT = model.LineVAT{
		CategoryCode:        l.Item.ClassifiedTaxCategory.ID,
		Rate:                model.Percentage(l.Item.ClassifiedTaxCategory.Percent),
		ExemptionReasonText: l.Item.ClassifiedTaxCategory.TaxExemptionReason,
		ExemptionReasonCode: l.Item.ClassifiedTaxCategory.TaxExemptionReasonCode,
	}
	for _, ac := range l.AllowanceCharge {
		lac := model.LineAllowanceCharge{
			Amount:     model.Amount(ac.Amount.Value),
			Percentage: model.Percentage(ac.MultiplierFactorNumeric),
			Reason:     ac.AllowanceChargeReason,
			ReasonCode: ac.AllowanceChargeReasonCode,
		}
		if ac.BaseAmount != nil {
			lac.BaseAmount = model.Amount(ac.BaseAmount.Value)
		}
		if ac.ChargeIndicator {
			ml.Charges = append(ml.Charges, lac)
		} else {
			ml.Allowances = append(ml.Allowances, lac)
		}
	}
	for _, sub := range l.SubInvoiceLine {
		ml.SubLines = append(ml.SubLines, toLine(sub))
	}
	return ml
}

func toItem(it Item) model.Item {
	mi := model.Item{
		Name:            it.Name,
		Description:     it.Description,
		CountryOfOrigin: "",
	}
	if it.SellersItemIdentification != nil {
		mi.SellerIdentifier = it.SellersItemIdentification.ID
	}
	if it.BuyersItemIdentification != nil {
		mi.BuyerIdentifier = it.BuyersItemIdentification.ID
	}
	if it.StandardItemIdentification != nil {
		mi.StandardIdentifier = &model.ID{Value: it.StandardItemIdentification.ID.Value, Scheme: it.StandardItemIdentification.ID.SchemeID}
	}
	if it.OriginCountry != nil {
		mi.CountryOfOrigin = it.OriginCountry.IdentificationCode
	}
	for _, cc := range it.CommodityClassification {
		mi.Classifications = append(mi.Classifications, model.ItemClassification{
			Code:        cc.ItemClassificationCode.Value,
			ListID:      cc.ItemClassificationCode.ListID,
			ListVersion: cc.ItemClassificationCode.ListVersion,
		})
	}
	for _, p := range it.AdditionalItemProperty {
		mi.Attributes = append(mi.Attributes, model.ItemAttribute{Name: p.Name, Value: p.Value})
	}
	return mi
}

func toPrice(p Price) model.Price {
	mp := model.Price{NetPrice: model.Amount(p.PriceAmount.Value)}
	if p.BaseQuantity != nil {
		mp.BaseQuantity = &model.Quantity{Value: model.Decimal(p.BaseQuantity.Value), Unit: p.BaseQuantity.Unit}
	}
	if ac := p.AllowanceCharge; ac != nil {
		mp.Discount = model.Amount(ac.Amount.Value)
		if ac.BaseAmount != nil {
			mp.GrossPrice = model.Amount(ac.BaseAmount.Value)
		}
	}
	return mp
}
