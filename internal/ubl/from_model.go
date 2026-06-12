package ubl

import (
	"encoding/xml"

	"github.com/andeedotnet/go-xinvoice/model"
)

// creditNoteTypeCodes are the UNTDID 1001 codes carried as a UBL CreditNote.
var creditNoteTypeCodes = map[string]bool{
	"81": true, "83": true, "261": true, "262": true, "296": true, "308": true,
	"381": true, "396": true, "420": true, "458": true, "532": true,
}

func isCreditNoteType(code string) bool { return creditNoteTypeCodes[code] }

// fromNote re-encodes a note's subject code (BT-21) as the UBL "#CODE#" prefix.
func fromNote(n model.Note) string {
	if n.SubjectCode != "" {
		return "#" + n.SubjectCode + "#" + n.Text
	}
	return n.Text
}

// fromModel converts the syntax-neutral model into a UBL [Invoice]. Amounts are
// emitted in the document currency (BT-5), except the accounting-currency VAT
// total (BT-111), which uses BT-6.
func fromModel(inv *model.Invoice) *Invoice {
	cur := inv.CurrencyCode
	creditNote := isCreditNoteType(inv.TypeCode)
	u := &Invoice{
		CustomizationID:      inv.SpecificationIdentifier,
		ProfileID:            inv.BusinessProcessType,
		ID:                   inv.Number,
		IssueDate:            string(inv.IssueDate),
		DueDate:              string(inv.DueDate),
		TaxPointDate:         string(inv.TaxPointDate),
		DocumentCurrencyCode: cur,
		TaxCurrencyCode:      inv.VATAccountingCurrency,
		AccountingCost:       inv.BuyerAccountingReference,
		BuyerReference:       inv.BuyerReference,
	}
	if creditNote {
		u.XMLName = xml.Name{Space: nsCreditNote, Local: "CreditNote"}
		u.CreditNoteTypeCode = inv.TypeCode
	} else {
		u.XMLName = xml.Name{Space: nsInvoice, Local: "Invoice"}
		u.InvoiceTypeCode = inv.TypeCode
	}
	for _, n := range inv.Notes {
		u.Note = append(u.Note, fromNote(n))
	}
	if inv.PurchaseOrderReference != "" || inv.SalesOrderReference != "" {
		u.OrderReference = &OrderReference{ID: inv.PurchaseOrderReference, SalesOrderID: inv.SalesOrderReference}
	}
	if inv.ContractReference != "" {
		u.ContractDocumentReference = &DocRef{ID: inv.ContractReference}
	}
	if inv.ProjectReference != "" {
		u.ProjectReference = &DocRef{ID: inv.ProjectReference}
	}
	if inv.DespatchAdviceReference != "" {
		u.DespatchDocumentReference = &DocRef{ID: inv.DespatchAdviceReference}
	}
	if inv.ReceivingAdviceReference != "" {
		u.ReceiptDocumentReference = &DocRef{ID: inv.ReceivingAdviceReference}
	}
	if inv.TenderOrLotReference != "" {
		u.OriginatorDocumentReference = &DocRef{ID: inv.TenderOrLotReference}
	}
	for _, pi := range inv.PrecedingInvoices {
		u.BillingReference = append(u.BillingReference, BillingReference{
			InvoiceDocumentReference: DocRefWithDate{ID: pi.Reference, IssueDate: string(pi.IssueDate)},
		})
	}
	if inv.InvoicedObject != nil {
		u.AdditionalDocumentReference = append(u.AdditionalDocumentReference, AdditionalDocRef{
			ID:               idCode{Value: inv.InvoicedObject.Value, SchemeID: inv.InvoicedObject.Scheme},
			DocumentTypeCode: "130",
		})
	}
	for _, d := range inv.Documents {
		ref := AdditionalDocRef{ID: idCode{Value: d.Reference}, DocumentDescription: d.Description}
		if d.ExternalLocation != "" {
			ref.Attachment = &Attachment{ExternalReference: &ExtRef{URI: d.ExternalLocation}}
		}
		if d.Attachment != nil {
			if ref.Attachment == nil {
				ref.Attachment = &Attachment{}
			}
			ref.Attachment.EmbeddedDocumentBinaryObject = &binary{
				Value:    string(d.Attachment.Content),
				MimeCode: d.Attachment.MimeCode,
				Filename: d.Attachment.Filename,
			}
		}
		u.AdditionalDocumentReference = append(u.AdditionalDocumentReference, ref)
	}

	u.AccountingSupplierParty = SupplierParty{Party: fromParty(inv.Seller)}
	u.AccountingCustomerParty = CustomerParty{Party: fromParty(inv.Buyer)}
	// BT-90 (bank assigned creditor identifier) is bound to a SEPA-schemed seller
	// party identification in UBL.
	if pi := inv.PaymentInstructions; pi != nil && pi.DirectDebit != nil && pi.DirectDebit.CreditorIdentifier != "" {
		u.AccountingSupplierParty.Party.PartyIdentification = append(
			u.AccountingSupplierParty.Party.PartyIdentification,
			PartyID{ID: idCode{Value: pi.DirectDebit.CreditorIdentifier, SchemeID: "SEPA"}})
	}
	if inv.Payee != nil {
		u.PayeeParty = fromPayee(inv.Payee)
	}
	if inv.TaxRepresentative != nil {
		u.TaxRepresentativeParty = fromTaxRep(inv.TaxRepresentative)
	}
	if d := inv.Delivery; d != nil {
		u.InvoicePeriod = fromPeriod(d.InvoicingPeriod)
		if d.ActualDeliveryDate != "" || d.PartyName != "" || d.LocationIdentifier != nil || d.Address != nil {
			u.Delivery = fromDelivery(d)
		}
	}
	u.PaymentMeans = fromPaymentInstructions(inv.PaymentInstructions)
	if inv.PaymentTerms != "" {
		u.PaymentTerms = &PaymentTerms{Note: inv.PaymentTerms}
	}
	for _, tp := range inv.ThirdPartyPayments {
		u.PrepaidPayment = append(u.PrepaidPayment, PrepaidPayment{
			ID:            tp.Type,
			PaidAmount:    amount{Value: string(tp.Amount), Currency: cur},
			InstructionID: tp.Description,
		})
	}
	for _, a := range inv.Allowances {
		u.AllowanceCharge = append(u.AllowanceCharge, fromAllowanceCharge(a, false, cur))
	}
	for _, c := range inv.Charges {
		u.AllowanceCharge = append(u.AllowanceCharge, fromAllowanceCharge(c, true, cur))
	}

	u.TaxTotal = fromTax(inv.VATBreakdown, inv.Totals.TaxTotal, inv.Totals.TaxTotalAccountingCurrency, cur, inv.VATAccountingCurrency)
	u.LegalMonetaryTotal = fromTotals(inv.Totals, cur)
	for _, l := range inv.Lines {
		line := fromLine(l, cur, creditNote)
		if creditNote {
			u.CreditNoteLine = append(u.CreditNoteLine, line)
		} else {
			u.InvoiceLine = append(u.InvoiceLine, line)
		}
	}
	return u
}

func fromParty(p model.Party) Party {
	up := Party{PostalAddress: fromAddress(p.Address), Contact: fromContact(p.Contact)}
	if p.ElectronicAddress != nil {
		up.EndpointID = &idCode{Value: p.ElectronicAddress.Value, SchemeID: p.ElectronicAddress.Scheme}
	}
	for _, id := range p.Identifiers {
		up.PartyIdentification = append(up.PartyIdentification, PartyID{ID: idCode{Value: id.Value, SchemeID: id.Scheme}})
	}
	if p.TradingName != "" {
		up.PartyName = &PartyName{Name: p.TradingName}
	}
	if p.VATIdentifier != "" {
		up.PartyTaxScheme = append(up.PartyTaxScheme, PartyTaxScheme{CompanyID: p.VATIdentifier, TaxScheme: TaxScheme{ID: "VAT"}})
	}
	if p.TaxRegistrationID != "" {
		up.PartyTaxScheme = append(up.PartyTaxScheme, PartyTaxScheme{CompanyID: p.TaxRegistrationID, TaxScheme: TaxScheme{ID: "FC"}})
	}
	if p.Name != "" || p.LegalRegistrationID != nil || p.AdditionalLegalInfo != "" {
		le := &LegalEntity{RegistrationName: p.Name, CompanyLegalForm: p.AdditionalLegalInfo}
		if p.LegalRegistrationID != nil {
			le.CompanyID = &idCode{Value: p.LegalRegistrationID.Value, SchemeID: p.LegalRegistrationID.Scheme}
		}
		up.PartyLegalEntity = le
	}
	return up
}

func fromPayee(p *model.Payee) *Party {
	up := &Party{}
	if p.Name != "" {
		up.PartyName = &PartyName{Name: p.Name}
	}
	if p.Identifier != nil {
		up.PartyIdentification = append(up.PartyIdentification, PartyID{ID: idCode{Value: p.Identifier.Value, SchemeID: p.Identifier.Scheme}})
	}
	if p.LegalRegistrationID != nil {
		up.PartyLegalEntity = &LegalEntity{CompanyID: &idCode{Value: p.LegalRegistrationID.Value, SchemeID: p.LegalRegistrationID.Scheme}}
	}
	return up
}

func fromTaxRep(t *model.TaxRepresentative) *Party {
	up := &Party{PostalAddress: fromAddress(t.Address)}
	if t.Name != "" {
		up.PartyName = &PartyName{Name: t.Name}
	}
	if t.VATIdentifier != "" {
		up.PartyTaxScheme = append(up.PartyTaxScheme, PartyTaxScheme{CompanyID: t.VATIdentifier, TaxScheme: TaxScheme{ID: "VAT"}})
	}
	return up
}

func fromAddress(a model.Address) *Address {
	if a == (model.Address{}) {
		return nil
	}
	ua := &Address{
		StreetName:           a.Line1,
		AdditionalStreetName: a.Line2,
		CityName:             a.City,
		PostalZone:           a.PostCode,
		CountrySubentity:     a.Subdivision,
		Country:              Country{IdentificationCode: a.CountryCode},
	}
	if a.Line3 != "" {
		ua.AddressLine = &AddressLine{Line: a.Line3}
	}
	return ua
}

func fromContact(c *model.Contact) *Contact {
	if c == nil {
		return nil
	}
	return &Contact{Name: c.Point, Telephone: c.Phone, ElectronicMail: c.Email}
}

func fromDelivery(d *model.Delivery) *Delivery {
	ud := &Delivery{ActualDeliveryDate: string(d.ActualDeliveryDate)}
	if d.LocationIdentifier != nil || d.Address != nil {
		loc := &DeliveryLocation{}
		if d.LocationIdentifier != nil {
			loc.ID = &idCode{Value: d.LocationIdentifier.Value, SchemeID: d.LocationIdentifier.Scheme}
		}
		if d.Address != nil {
			loc.Address = fromAddress(*d.Address)
		}
		ud.DeliveryLocation = loc
	}
	if d.PartyName != "" {
		ud.DeliveryParty = &Party{PartyName: &PartyName{Name: d.PartyName}}
	}
	return ud
}

func fromPeriod(p *model.Period) *Period {
	if p == nil {
		return nil
	}
	return &Period{StartDate: string(p.Start), EndDate: string(p.End)}
}

func fromPaymentInstructions(pi *model.PaymentInstructions) []PaymentMeans {
	if pi == nil {
		return nil
	}
	code := typedCode{Value: pi.MeansTypeCode, Name: pi.MeansText}
	var means []PaymentMeans
	emitID := pi.RemittanceInformation
	for _, ct := range pi.CreditTransfers {
		fa := &FinancialAccount{ID: ct.AccountIdentifier, Name: ct.AccountName}
		if ct.ServiceProviderID != "" {
			fa.FinancialInstitutionBranch = &FinancialInstitutionBranch{ID: ct.ServiceProviderID}
		}
		means = append(means, PaymentMeans{PaymentMeansCode: code, PaymentID: emitID, PayeeFinancialAccount: fa})
		emitID = ""
	}
	if pi.Card != nil {
		means = append(means, PaymentMeans{PaymentMeansCode: code, PaymentID: emitID, CardAccount: &CardAccount{
			PrimaryAccountNumberID: pi.Card.PrimaryAccountNumber, HolderName: pi.Card.HolderName,
		}})
		emitID = ""
	}
	if pi.DirectDebit != nil {
		pm := &PaymentMandate{ID: pi.DirectDebit.MandateReference}
		if pi.DirectDebit.DebitedAccount != "" {
			pm.PayerFinancialAccount = &FinancialAccount{ID: pi.DirectDebit.DebitedAccount}
		}
		means = append(means, PaymentMeans{PaymentMeansCode: code, PaymentID: emitID, PaymentMandate: pm})
		emitID = ""
	}
	if len(means) == 0 {
		means = append(means, PaymentMeans{PaymentMeansCode: code, PaymentID: emitID})
	}
	return means
}

func fromAllowanceCharge(a model.AllowanceCharge, charge bool, cur string) AllowanceCharge {
	ac := AllowanceCharge{
		ChargeIndicator:           charge,
		AllowanceChargeReasonCode: a.ReasonCode,
		AllowanceChargeReason:     a.Reason,
		MultiplierFactorNumeric:   string(a.Percentage),
		Amount:                    amount{Value: string(a.Amount), Currency: cur},
	}
	if a.BaseAmount != "" {
		ac.BaseAmount = &amount{Value: string(a.BaseAmount), Currency: cur}
	}
	if a.VATCategoryCode != "" || a.VATRate != "" {
		ac.TaxCategory = &TaxCategory{ID: a.VATCategoryCode, Percent: string(a.VATRate), TaxScheme: TaxScheme{ID: "VAT"}}
	}
	return ac
}

func fromTax(breakdown []model.VATBreakdown, taxTotal, taxTotalAcct model.Amount, cur, acctCur string) []TaxTotal {
	var totals []TaxTotal
	tt := TaxTotal{TaxAmount: amount{Value: string(taxTotal), Currency: cur}}
	for _, b := range breakdown {
		tc := TaxCategory{
			ID:                     b.CategoryCode,
			Percent:                string(b.Rate),
			TaxExemptionReasonCode: b.ExemptionReasonCode,
			TaxExemptionReason:     b.ExemptionReasonText,
			TaxScheme:              TaxScheme{ID: "VAT"},
		}
		tt.TaxSubtotal = append(tt.TaxSubtotal, TaxSubtotal{
			TaxableAmount: amount{Value: string(b.TaxableAmount), Currency: cur},
			TaxAmount:     amount{Value: string(b.TaxAmount), Currency: cur},
			TaxCategory:   tc,
		})
	}
	totals = append(totals, tt)
	if taxTotalAcct != "" {
		totals = append(totals, TaxTotal{TaxAmount: amount{Value: string(taxTotalAcct), Currency: acctCur}})
	}
	return totals
}

func fromTotals(t model.DocumentTotals, cur string) MonetaryTotal {
	mt := MonetaryTotal{
		LineExtensionAmount: amount{Value: string(t.LineNetTotal), Currency: cur},
		TaxExclusiveAmount:  amount{Value: string(t.TaxBasisTotal), Currency: cur},
		TaxInclusiveAmount:  amount{Value: string(t.GrandTotal), Currency: cur},
		PayableAmount:       amount{Value: string(t.DuePayableAmount), Currency: cur},
	}
	if t.AllowanceTotal != "" {
		mt.AllowanceTotalAmount = &amount{Value: string(t.AllowanceTotal), Currency: cur}
	}
	if t.ChargeTotal != "" {
		mt.ChargeTotalAmount = &amount{Value: string(t.ChargeTotal), Currency: cur}
	}
	if t.PaidAmount != "" {
		mt.PrepaidAmount = &amount{Value: string(t.PaidAmount), Currency: cur}
	}
	if t.RoundingAmount != "" {
		mt.PayableRoundingAmount = &amount{Value: string(t.RoundingAmount), Currency: cur}
	}
	return mt
}

func fromLine(l model.Line, cur string, creditNote bool) InvoiceLine {
	q := &quantity{Value: string(l.Quantity.Value), Unit: l.Quantity.Unit}
	ul := InvoiceLine{
		ID:                  l.ID,
		Note:                l.Note,
		LineExtensionAmount: amount{Value: string(l.NetAmount), Currency: cur},
		AccountingCost:      l.BuyerAccountingReference,
		InvoicePeriod:       fromPeriod(l.Period),
		Item:                fromItem(l.Item, l.VAT),
		Price:               fromPrice(l.Price, cur),
	}
	if creditNote {
		ul.CreditedQuantity = q
	} else {
		ul.InvoicedQuantity = q
	}
	if l.OrderLineReference != "" {
		ul.OrderLineReference = &OrderLineRef{LineID: l.OrderLineReference}
	}
	if l.ObjectIdentifier != nil {
		ul.DocumentReference = &LineDocRef{
			ID:               idCode{Value: l.ObjectIdentifier.Value, SchemeID: l.ObjectIdentifier.Scheme},
			DocumentTypeCode: "130",
		}
	}
	for _, a := range l.Allowances {
		ul.AllowanceCharge = append(ul.AllowanceCharge, fromLineAllowance(a, false, cur))
	}
	for _, c := range l.Charges {
		ul.AllowanceCharge = append(ul.AllowanceCharge, fromLineAllowance(c, true, cur))
	}
	for _, sub := range l.SubLines {
		ul.SubInvoiceLine = append(ul.SubInvoiceLine, fromLine(sub, cur, creditNote))
	}
	return ul
}

func fromLineAllowance(a model.LineAllowanceCharge, charge bool, cur string) AllowanceCharge {
	ac := AllowanceCharge{
		ChargeIndicator:           charge,
		AllowanceChargeReasonCode: a.ReasonCode,
		AllowanceChargeReason:     a.Reason,
		MultiplierFactorNumeric:   string(a.Percentage),
		Amount:                    amount{Value: string(a.Amount), Currency: cur},
	}
	if a.BaseAmount != "" {
		ac.BaseAmount = &amount{Value: string(a.BaseAmount), Currency: cur}
	}
	return ac
}

func fromItem(it model.Item, vat model.LineVAT) Item {
	ui := Item{
		Description: it.Description,
		Name:        it.Name,
		ClassifiedTaxCategory: TaxCategory{
			ID:                     vat.CategoryCode,
			Percent:                string(vat.Rate),
			TaxExemptionReasonCode: vat.ExemptionReasonCode,
			TaxExemptionReason:     vat.ExemptionReasonText,
			TaxScheme:              TaxScheme{ID: "VAT"},
		},
	}
	if it.BuyerIdentifier != "" {
		ui.BuyersItemIdentification = &ItemID{ID: it.BuyerIdentifier}
	}
	if it.SellerIdentifier != "" {
		ui.SellersItemIdentification = &ItemID{ID: it.SellerIdentifier}
	}
	if it.StandardIdentifier != nil {
		ui.StandardItemIdentification = &StandardItemID{ID: idCode{Value: it.StandardIdentifier.Value, SchemeID: it.StandardIdentifier.Scheme}}
	}
	if it.CountryOfOrigin != "" {
		ui.OriginCountry = &Country{IdentificationCode: it.CountryOfOrigin}
	}
	for _, c := range it.Classifications {
		ui.CommodityClassification = append(ui.CommodityClassification, CommodityClass{
			ItemClassificationCode: listCode{Value: c.Code, ListID: c.ListID, ListVersion: c.ListVersion},
		})
	}
	for _, p := range it.Attributes {
		ui.AdditionalItemProperty = append(ui.AdditionalItemProperty, ItemProperty{Name: p.Name, Value: p.Value})
	}
	return ui
}

func fromPrice(p model.Price, cur string) Price {
	up := Price{PriceAmount: amount{Value: string(p.NetPrice), Currency: cur}}
	if p.BaseQuantity != nil {
		up.BaseQuantity = &quantity{Value: string(p.BaseQuantity.Value), Unit: p.BaseQuantity.Unit}
	}
	if p.Discount != "" || p.GrossPrice != "" {
		pa := &PriceAllowance{ChargeIndicator: false, Amount: amount{Value: string(p.Discount), Currency: cur}}
		if p.GrossPrice != "" {
			pa.BaseAmount = &amount{Value: string(p.GrossPrice), Currency: cur}
		}
		up.AllowanceCharge = pa
	}
	return up
}
