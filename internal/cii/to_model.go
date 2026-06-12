package cii

import "github.com/andeedotnet/go-xinvoice/model"

// toModel converts a parsed CII [Invoice] into the syntax-neutral model.
func toModel(c *Invoice) *model.Invoice {
	ctx := c.Context
	doc := c.Document
	ag := c.Transaction.Agreement
	dl := c.Transaction.Delivery
	st := c.Transaction.Settlement

	inv := &model.Invoice{
		SpecificationIdentifier: ctx.Guideline.ID,
		Number:                  doc.ID,
		TypeCode:                doc.TypeCode,
		IssueDate:               dateOf(doc.IssueDateTime),
		CurrencyCode:            st.InvoiceCurrencyCode,
		VATAccountingCurrency:   st.TaxCurrencyCode,
		BuyerReference:          ag.BuyerReference,
		Seller:                  toParty(&ag.SellerTradeParty),
		Buyer:                   toParty(&ag.BuyerTradeParty),
	}
	if ctx.BusinessProcess != nil {
		inv.BusinessProcessType = ctx.BusinessProcess.ID
	}
	for _, n := range doc.IncludedNote {
		inv.Notes = append(inv.Notes, model.Note{Text: n.Content, SubjectCode: n.SubjectCode})
	}

	if ag.BuyerOrderReferencedDoc != nil {
		inv.PurchaseOrderReference = ag.BuyerOrderReferencedDoc.IssuerAssignedID
	}
	if ag.SellerOrderReferencedDoc != nil {
		inv.SalesOrderReference = ag.SellerOrderReferencedDoc.IssuerAssignedID
	}
	if ag.ContractReferencedDocument != nil {
		inv.ContractReference = ag.ContractReferencedDocument.IssuerAssignedID
	}
	if ag.ProcuringProject != nil {
		inv.ProjectReference = ag.ProcuringProject.ID
	}
	if dl.DespatchAdviceRef != nil {
		inv.DespatchAdviceReference = dl.DespatchAdviceRef.IssuerAssignedID
	}
	if dl.ReceivingAdviceRef != nil {
		inv.ReceivingAdviceReference = dl.ReceivingAdviceRef.IssuerAssignedID
	}
	for _, d := range ag.AdditionalReferencedDoc {
		switch d.TypeCode {
		case "130":
			inv.InvoicedObject = &model.ID{Value: d.IssuerAssignedID, Scheme: d.ReferenceTypeCode}
		case "50":
			inv.TenderOrLotReference = d.IssuerAssignedID
		default:
			sd := model.SupportingDocument{
				Reference:        d.IssuerAssignedID,
				Description:      d.Name,
				ExternalLocation: d.URIID,
			}
			if d.AttachmentBinary != nil {
				sd.Attachment = &model.BinaryObject{
					MimeCode: d.AttachmentBinary.MimeCode,
					Filename: d.AttachmentBinary.Filename,
					Content:  []byte(d.AttachmentBinary.Value),
				}
			}
			inv.Documents = append(inv.Documents, sd)
		}
	}

	if st.AccountingAccount != nil {
		inv.BuyerAccountingReference = st.AccountingAccount.ID
	}
	if st.PaymentReference != "" {
		// carried into payment instructions below
	}
	for _, r := range st.InvoiceReferencedDoc {
		inv.PrecedingInvoices = append(inv.PrecedingInvoices, model.PrecedingInvoiceReference{
			Reference: r.IssuerAssignedID,
			IssueDate: qdtDateOf(r.FormattedIssueDateTime),
		})
	}

	if st.PayeeTradeParty != nil {
		inv.Payee = toPayee(st.PayeeTradeParty)
	}
	if ag.TaxRepresentativeParty != nil {
		inv.TaxRepresentative = toTaxRep(ag.TaxRepresentativeParty)
	}
	if d := toDelivery(&dl, st.BillingPeriod); d != nil {
		inv.Delivery = d
	}
	inv.PaymentInstructions = toPaymentInstructions(&st)
	if pt := st.PaymentTerms; pt != nil {
		inv.PaymentTerms = pt.Description
		inv.DueDate = dateOf(pt.DueDateDateTime)
	}

	for _, ac := range st.AllowanceCharge {
		m := toAllowanceCharge(ac)
		if ac.ChargeIndicator.Indicator {
			inv.Charges = append(inv.Charges, m)
		} else {
			inv.Allowances = append(inv.Allowances, m)
		}
	}

	inv.VATBreakdown = toVATBreakdown(st.ApplicableTradeTax)
	inv.Totals = toTotals(st.Summation)

	// BT-7 / BT-8: the VAT point date / code lives per trade tax in CII but is a
	// single document-level value in the model; take the first occurrence.
	for i := range st.ApplicableTradeTax {
		t := &st.ApplicableTradeTax[i]
		if t.TaxPointDate != nil && inv.TaxPointDate == "" {
			inv.TaxPointDate = model.Date(fromCIIDate(t.TaxPointDate.DateString.Value))
		}
		if t.DueDateTypeCode != "" && inv.TaxPointDateCode == "" {
			inv.TaxPointDateCode = t.DueDateTypeCode
		}
	}

	for _, l := range c.Transaction.Lines {
		inv.Lines = append(inv.Lines, toLine(l))
	}
	return inv
}

func toParty(p *Party) model.Party {
	mp := model.Party{
		Name:                p.Name,
		AdditionalLegalInfo: p.Description, // BT-33 (seller free-text legal info)
		Address:             toAddress(p.PostalAddress),
		Contact:             toContact(p.Contact),
	}
	for _, id := range p.ID {
		mp.Identifiers = append(mp.Identifiers, model.ID{Value: id})
	}
	for _, g := range p.GlobalID {
		mp.Identifiers = append(mp.Identifiers, model.ID{Value: g.Value, Scheme: g.SchemeID})
	}
	if lo := p.LegalOrg; lo != nil {
		mp.TradingName = lo.TradingBusinessName
		if lo.ID != nil {
			mp.LegalRegistrationID = &model.ID{Value: lo.ID.Value, Scheme: lo.ID.SchemeID}
		}
	}
	if p.URIComm != nil {
		mp.ElectronicAddress = &model.ID{Value: p.URIComm.URIID.Value, Scheme: p.URIComm.URIID.SchemeID}
	}
	for _, tr := range p.TaxRegistration {
		if tr.ID.SchemeID == "VA" {
			mp.VATIdentifier = tr.ID.Value
		} else {
			mp.TaxRegistrationID = tr.ID.Value
		}
	}
	return mp
}

func toPayee(p *Party) *model.Payee {
	pe := &model.Payee{Name: p.Name}
	if len(p.ID) > 0 {
		pe.Identifier = &model.ID{Value: p.ID[0]}
	} else if len(p.GlobalID) > 0 {
		pe.Identifier = &model.ID{Value: p.GlobalID[0].Value, Scheme: p.GlobalID[0].SchemeID}
	}
	if p.LegalOrg != nil && p.LegalOrg.ID != nil {
		pe.LegalRegistrationID = &model.ID{Value: p.LegalOrg.ID.Value, Scheme: p.LegalOrg.ID.SchemeID}
	}
	return pe
}

func toTaxRep(p *Party) *model.TaxRepresentative {
	tr := &model.TaxRepresentative{Name: p.Name, Address: toAddress(p.PostalAddress)}
	for _, r := range p.TaxRegistration {
		if r.ID.SchemeID == "VA" {
			tr.VATIdentifier = r.ID.Value
		}
	}
	return tr
}

func toAddress(a *Address) model.Address {
	if a == nil {
		return model.Address{}
	}
	return model.Address{
		Line1:       a.LineOne,
		Line2:       a.LineTwo,
		Line3:       a.LineThree,
		City:        a.CityName,
		PostCode:    a.PostcodeCode,
		Subdivision: a.CountrySubDivision,
		CountryCode: a.CountryID,
	}
}

func toContact(c *Contact) *model.Contact {
	if c == nil {
		return nil
	}
	mc := &model.Contact{Point: c.PersonName}
	if mc.Point == "" {
		mc.Point = c.DepartmentName
	}
	if c.Telephone != nil {
		mc.Phone = c.Telephone.CompleteNumber
	}
	if c.Email != nil {
		mc.Email = c.Email.URIID
	}
	return mc
}

func toDelivery(d *Delivery, period *period) *model.Delivery {
	md := &model.Delivery{InvoicingPeriod: toPeriod(period)}
	has := md.InvoicingPeriod != nil
	if d.ActualDelivery != nil {
		md.ActualDeliveryDate = dateOf(d.ActualDelivery.OccurrenceDateTime)
		has = has || md.ActualDeliveryDate != ""
	}
	if sp := d.ShipToTradeParty; sp != nil {
		md.PartyName = sp.Name
		if len(sp.ID) > 0 {
			md.LocationIdentifier = &model.ID{Value: sp.ID[0]}
		} else if len(sp.GlobalID) > 0 {
			md.LocationIdentifier = &model.ID{Value: sp.GlobalID[0].Value, Scheme: sp.GlobalID[0].SchemeID}
		}
		if sp.PostalAddress != nil {
			a := toAddress(sp.PostalAddress)
			md.Address = &a
		}
		has = true
	}
	if !has {
		return nil
	}
	return md
}

func toPeriod(p *period) *model.Period {
	if p == nil {
		return nil
	}
	start := dateOf(p.StartDateTime)
	end := dateOf(p.EndDateTime)
	if start == "" && end == "" {
		return nil
	}
	return &model.Period{Start: start, End: end}
}

func toPaymentInstructions(st *Settlement) *model.PaymentInstructions {
	hasMandate := st.PaymentTerms != nil && st.PaymentTerms.DirectDebitMandateID != ""
	if len(st.PaymentMeans) == 0 && st.PaymentReference == "" && st.CreditorReferenceID == "" && !hasMandate {
		return nil
	}
	pi := &model.PaymentInstructions{RemittanceInformation: st.PaymentReference}
	for i, m := range st.PaymentMeans {
		if i == 0 {
			pi.MeansTypeCode = m.TypeCode
			pi.MeansText = m.Information
		}
		if a := m.PayeeAccount; a != nil {
			id := a.IBANID
			if id == "" {
				id = a.ProprietaryID
			}
			ct := model.CreditTransfer{AccountIdentifier: id, AccountName: a.AccountName}
			if m.PayeeInstitution != nil {
				ct.ServiceProviderID = m.PayeeInstitution.BICID
			}
			pi.CreditTransfers = append(pi.CreditTransfers, ct)
		}
		if fc := m.FinancialCard; fc != nil && pi.Card == nil {
			pi.Card = &model.PaymentCard{PrimaryAccountNumber: fc.ID, HolderName: fc.CardholderName}
		}
		if m.PayerAccount != nil && pi.DirectDebit == nil {
			pi.DirectDebit = &model.DirectDebit{DebitedAccount: m.PayerAccount.IBANID}
		}
	}
	if hasMandate || st.CreditorReferenceID != "" {
		if pi.DirectDebit == nil {
			pi.DirectDebit = &model.DirectDebit{}
		}
		if st.PaymentTerms != nil {
			pi.DirectDebit.MandateReference = st.PaymentTerms.DirectDebitMandateID
		}
		pi.DirectDebit.CreditorIdentifier = st.CreditorReferenceID
	}
	return pi
}

func toAllowanceCharge(ac AllowanceCharge) model.AllowanceCharge {
	m := model.AllowanceCharge{
		Amount:     model.Amount(ac.ActualAmount),
		BaseAmount: model.Amount(ac.BasisAmount),
		Percentage: model.Percentage(ac.CalculationPercent),
		Reason:     ac.Reason,
		ReasonCode: ac.ReasonCode,
	}
	if ct := ac.CategoryTradeTax; ct != nil {
		m.VATCategoryCode = ct.CategoryCode
		m.VATRate = model.Percentage(ct.RateApplicablePercent)
	}
	return m
}

func toVATBreakdown(taxes []TradeTax) []model.VATBreakdown {
	var out []model.VATBreakdown
	for _, t := range taxes {
		out = append(out, model.VATBreakdown{
			TaxableAmount:       model.Amount(t.BasisAmount),
			TaxAmount:           model.Amount(t.CalculatedAmount),
			CategoryCode:        t.CategoryCode,
			Rate:                model.Percentage(t.RateApplicablePercent),
			ExemptionReasonText: t.ExemptionReason,
			ExemptionReasonCode: t.ExemptionReasonCode,
		})
	}
	return out
}

func toTotals(s Summation) model.DocumentTotals {
	dt := model.DocumentTotals{
		LineNetTotal:     model.Amount(s.LineTotalAmount),
		AllowanceTotal:   model.Amount(s.AllowanceTotalAmount),
		ChargeTotal:      model.Amount(s.ChargeTotalAmount),
		TaxBasisTotal:    model.Amount(s.TaxBasisTotalAmount),
		GrandTotal:       model.Amount(s.GrandTotalAmount),
		PaidAmount:       model.Amount(s.TotalPrepaidAmount),
		RoundingAmount:   model.Amount(s.RoundingAmount),
		DuePayableAmount: model.Amount(s.DuePayableAmount),
	}
	if len(s.TaxTotalAmount) > 0 {
		dt.TaxTotal = model.Amount(s.TaxTotalAmount[0].Value)
	}
	if len(s.TaxTotalAmount) > 1 {
		dt.TaxTotalAccountingCurrency = model.Amount(s.TaxTotalAmount[1].Value)
	}
	return dt
}

func toLine(l LineItem) model.Line {
	ml := model.Line{
		ID:        l.LineDocument.LineID,
		Quantity:  model.Quantity{Value: model.Decimal(l.Delivery.BilledQuantity.Value), Unit: l.Delivery.BilledQuantity.Unit},
		NetAmount: model.Amount(l.Settlement.Summation.LineTotalAmount),
		Period:    toPeriod(l.Settlement.BillingPeriod),
		Item:      toItem(l.Product),
		Price:     toPrice(l.Agreement),
		VAT: model.LineVAT{
			CategoryCode:        l.Settlement.ApplicableTradeTax.CategoryCode,
			Rate:                model.Percentage(l.Settlement.ApplicableTradeTax.RateApplicablePercent),
			ExemptionReasonText: l.Settlement.ApplicableTradeTax.ExemptionReason,
			ExemptionReasonCode: l.Settlement.ApplicableTradeTax.ExemptionReasonCode,
		},
	}
	if l.LineDocument.IncludedNote != nil {
		ml.Note = l.LineDocument.IncludedNote.Content
	}
	if l.Agreement.BuyerOrderReferencedDocument != nil {
		ml.OrderLineReference = l.Agreement.BuyerOrderReferencedDocument.LineID
	}
	if l.Settlement.AccountingAccount != nil {
		ml.BuyerAccountingReference = l.Settlement.AccountingAccount.ID
	}
	if od := l.Settlement.ObjectDocument; od != nil {
		ml.ObjectIdentifier = &model.ID{Value: od.IssuerAssignedID, Scheme: od.ReferenceTypeCode}
	}
	for _, ac := range l.Settlement.AllowanceCharge {
		lac := model.LineAllowanceCharge{
			Amount:     model.Amount(ac.ActualAmount),
			BaseAmount: model.Amount(ac.BasisAmount),
			Percentage: model.Percentage(ac.CalculationPercent),
			Reason:     ac.Reason,
			ReasonCode: ac.ReasonCode,
		}
		if ac.ChargeIndicator.Indicator {
			ml.Charges = append(ml.Charges, lac)
		} else {
			ml.Allowances = append(ml.Allowances, lac)
		}
	}
	return ml
}

func toItem(p Product) model.Item {
	mi := model.Item{
		Name:             p.Name,
		Description:      p.Description,
		SellerIdentifier: p.SellerAssignedID,
		BuyerIdentifier:  p.BuyerAssignedID,
	}
	if p.GlobalID != nil {
		mi.StandardIdentifier = &model.ID{Value: p.GlobalID.Value, Scheme: p.GlobalID.SchemeID}
	}
	if p.OriginCountry != nil {
		mi.CountryOfOrigin = p.OriginCountry.ID
	}
	for _, c := range p.Classification {
		mi.Classifications = append(mi.Classifications, model.ItemClassification{
			Code:        c.ClassCode.Value,
			ListID:      c.ClassCode.ListID,
			ListVersion: c.ClassCode.ListVersion,
		})
	}
	for _, ch := range p.Characteristic {
		mi.Attributes = append(mi.Attributes, model.ItemAttribute{Name: ch.Description, Value: ch.Value})
	}
	return mi
}

func toPrice(a LineAgreement) model.Price {
	mp := model.Price{NetPrice: model.Amount(a.NetPrice.ChargeAmount)}
	if a.NetPrice.BasisQuantity != nil {
		mp.BaseQuantity = &model.Quantity{Value: model.Decimal(a.NetPrice.BasisQuantity.Value), Unit: a.NetPrice.BasisQuantity.Unit}
	}
	if g := a.GrossPrice; g != nil {
		mp.GrossPrice = model.Amount(g.ChargeAmount)
		if g.Discount != nil {
			mp.Discount = model.Amount(g.Discount.ActualAmount)
		}
		// BT-149/150 may be carried on the gross price when absent from the net.
		if mp.BaseQuantity == nil && g.BasisQuantity != nil {
			mp.BaseQuantity = &model.Quantity{Value: model.Decimal(g.BasisQuantity.Value), Unit: g.BasisQuantity.Unit}
		}
	}
	return mp
}

// dateOf converts a udt date wrapper to a model date (YYYY-MM-DD).
func dateOf(d *dateWrap) model.Date {
	if d == nil {
		return ""
	}
	return model.Date(fromCIIDate(d.DateTimeString.Value))
}

func qdtDateOf(d *qdtDateWrap) model.Date {
	if d == nil {
		return ""
	}
	return model.Date(fromCIIDate(d.DateTimeString.Value))
}

// fromCIIDate turns "20160621" into "2016-06-21". Other forms pass through.
func fromCIIDate(s string) string {
	if len(s) == 8 {
		return s[0:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}
