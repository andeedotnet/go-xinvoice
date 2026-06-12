package cii

import (
	"strings"

	"github.com/andeedotnet/go-xinvoice/model"
)

// fromModel converts the syntax-neutral model into a CII [Invoice].
func fromModel(inv *model.Invoice) *Invoice {
	cur := inv.CurrencyCode
	c := &Invoice{}
	c.Context.Guideline = paramID{ID: inv.SpecificationIdentifier}
	if inv.BusinessProcessType != "" {
		c.Context.BusinessProcess = &paramID{ID: inv.BusinessProcessType}
	}

	c.Document = ExchangedDoc{
		ID:            inv.Number,
		TypeCode:      inv.TypeCode,
		IssueDateTime: dateWrapOf(inv.IssueDate),
	}
	for _, n := range inv.Notes {
		c.Document.IncludedNote = append(c.Document.IncludedNote, Note{Content: n.Text, SubjectCode: n.SubjectCode})
	}

	ag := &c.Transaction.Agreement
	ag.BuyerReference = inv.BuyerReference
	ag.SellerTradeParty = fromParty(inv.Seller)
	ag.BuyerTradeParty = fromParty(inv.Buyer)
	if inv.TaxRepresentative != nil {
		ag.TaxRepresentativeParty = fromTaxRep(inv.TaxRepresentative)
	}
	if inv.PurchaseOrderReference != "" {
		ag.BuyerOrderReferencedDoc = &docRef{IssuerAssignedID: inv.PurchaseOrderReference}
	}
	if inv.SalesOrderReference != "" {
		ag.SellerOrderReferencedDoc = &docRef{IssuerAssignedID: inv.SalesOrderReference}
	}
	if inv.ContractReference != "" {
		ag.ContractReferencedDocument = &docRef{IssuerAssignedID: inv.ContractReference}
	}
	if inv.ProjectReference != "" {
		ag.ProcuringProject = &project{ID: inv.ProjectReference, Name: "Projekt"}
	}
	if inv.InvoicedObject != nil {
		ag.AdditionalReferencedDoc = append(ag.AdditionalReferencedDoc, AddlDoc{
			IssuerAssignedID:  inv.InvoicedObject.Value,
			TypeCode:          "130",
			ReferenceTypeCode: inv.InvoicedObject.Scheme,
		})
	}
	if inv.TenderOrLotReference != "" {
		ag.AdditionalReferencedDoc = append(ag.AdditionalReferencedDoc, AddlDoc{
			IssuerAssignedID: inv.TenderOrLotReference,
			TypeCode:         "50",
		})
	}
	for _, d := range inv.Documents {
		ad := AddlDoc{IssuerAssignedID: d.Reference, TypeCode: "916", Name: d.Description, URIID: d.ExternalLocation}
		if d.Attachment != nil {
			ad.AttachmentBinary = &binary{Value: string(d.Attachment.Content), MimeCode: d.Attachment.MimeCode, Filename: d.Attachment.Filename}
		}
		ag.AdditionalReferencedDoc = append(ag.AdditionalReferencedDoc, ad)
	}

	dl := &c.Transaction.Delivery
	if d := inv.Delivery; d != nil {
		if d.ActualDeliveryDate != "" {
			dl.ActualDelivery = &deliveryEvent{OccurrenceDateTime: dateWrapOf(d.ActualDeliveryDate)}
		}
		if d.PartyName != "" || d.LocationIdentifier != nil || d.Address != nil {
			sp := &Party{Name: d.PartyName}
			if d.LocationIdentifier != nil {
				if d.LocationIdentifier.Scheme == "" {
					sp.ID = []string{d.LocationIdentifier.Value}
				} else {
					sp.GlobalID = []idCode{{Value: d.LocationIdentifier.Value, SchemeID: d.LocationIdentifier.Scheme}}
				}
			}
			if d.Address != nil {
				sp.PostalAddress = fromAddress(*d.Address)
			}
			dl.ShipToTradeParty = sp
		}
	}
	if inv.DespatchAdviceReference != "" {
		dl.DespatchAdviceRef = &docRef{IssuerAssignedID: inv.DespatchAdviceReference}
	}
	if inv.ReceivingAdviceReference != "" {
		dl.ReceivingAdviceRef = &docRef{IssuerAssignedID: inv.ReceivingAdviceReference}
	}

	st := &c.Transaction.Settlement
	st.InvoiceCurrencyCode = cur
	st.TaxCurrencyCode = inv.VATAccountingCurrency
	if inv.Payee != nil {
		st.PayeeTradeParty = fromPayee(inv.Payee)
	}
	fromPaymentInstructions(inv, st)
	for _, b := range inv.VATBreakdown {
		st.ApplicableTradeTax = append(st.ApplicableTradeTax, TradeTax{
			CalculatedAmount:      string(b.TaxAmount),
			TypeCode:              "VAT",
			ExemptionReason:       b.ExemptionReasonText,
			BasisAmount:           string(b.TaxableAmount),
			CategoryCode:          b.CategoryCode,
			ExemptionReasonCode:   b.ExemptionReasonCode,
			RateApplicablePercent: string(b.Rate),
		})
	}
	// BT-7 / BT-8 attach to the first trade tax (mutually exclusive per BR-CO-3).
	if len(st.ApplicableTradeTax) > 0 {
		if inv.TaxPointDate != "" {
			st.ApplicableTradeTax[0].TaxPointDate = &dateOnlyWrap{DateString: dateStr{Format: "102", Value: toCIIDate(string(inv.TaxPointDate))}}
		} else if inv.TaxPointDateCode != "" {
			st.ApplicableTradeTax[0].DueDateTypeCode = inv.TaxPointDateCode
		}
	}
	if inv.Delivery != nil {
		st.BillingPeriod = fromPeriod(inv.Delivery.InvoicingPeriod)
	}
	for _, a := range inv.Allowances {
		st.AllowanceCharge = append(st.AllowanceCharge, fromAllowanceCharge(a, false))
	}
	for _, ch := range inv.Charges {
		st.AllowanceCharge = append(st.AllowanceCharge, fromAllowanceCharge(ch, true))
	}
	st.PaymentTerms = fromPaymentTerms(inv)
	st.Summation = fromTotals(inv.Totals, cur, inv.VATAccountingCurrency)
	for _, pi := range inv.PrecedingInvoices {
		st.InvoiceReferencedDoc = append(st.InvoiceReferencedDoc, InvoiceRef{
			IssuerAssignedID:       pi.Reference,
			FormattedIssueDateTime: qdtDateWrapOf(pi.IssueDate),
		})
	}
	if inv.BuyerAccountingReference != "" {
		st.AccountingAccount = &account{ID: inv.BuyerAccountingReference}
	}

	for _, l := range inv.Lines {
		c.Transaction.Lines = append(c.Transaction.Lines, fromLine(l))
	}
	return c
}

func fromParty(p model.Party) Party {
	cp := Party{Name: p.Name, Description: p.AdditionalLegalInfo}
	for _, id := range p.Identifiers {
		if id.Scheme == "" {
			cp.ID = append(cp.ID, id.Value)
		} else {
			cp.GlobalID = append(cp.GlobalID, idCode{Value: id.Value, SchemeID: id.Scheme})
		}
	}
	if p.TradingName != "" || p.LegalRegistrationID != nil {
		lo := &LegalOrg{TradingBusinessName: p.TradingName}
		if p.LegalRegistrationID != nil {
			lo.ID = &idCode{Value: p.LegalRegistrationID.Value, SchemeID: p.LegalRegistrationID.Scheme}
		}
		cp.LegalOrg = lo
	}
	cp.Contact = fromContact(p.Contact)
	cp.PostalAddress = fromAddress(p.Address)
	if p.ElectronicAddress != nil {
		cp.URIComm = &uriComm{URIID: idCode{Value: p.ElectronicAddress.Value, SchemeID: p.ElectronicAddress.Scheme}}
	}
	if p.VATIdentifier != "" {
		cp.TaxRegistration = append(cp.TaxRegistration, TaxRegistration{ID: idCode{Value: p.VATIdentifier, SchemeID: "VA"}})
	}
	if p.TaxRegistrationID != "" {
		cp.TaxRegistration = append(cp.TaxRegistration, TaxRegistration{ID: idCode{Value: p.TaxRegistrationID, SchemeID: "FC"}})
	}
	return cp
}

func fromPayee(p *model.Payee) *Party {
	cp := &Party{Name: p.Name}
	if p.Identifier != nil {
		if p.Identifier.Scheme == "" {
			cp.ID = []string{p.Identifier.Value}
		} else {
			cp.GlobalID = []idCode{{Value: p.Identifier.Value, SchemeID: p.Identifier.Scheme}}
		}
	}
	if p.LegalRegistrationID != nil {
		cp.LegalOrg = &LegalOrg{ID: &idCode{Value: p.LegalRegistrationID.Value, SchemeID: p.LegalRegistrationID.Scheme}}
	}
	return cp
}

func fromTaxRep(t *model.TaxRepresentative) *Party {
	cp := &Party{Name: t.Name, PostalAddress: fromAddress(t.Address)}
	if t.VATIdentifier != "" {
		cp.TaxRegistration = append(cp.TaxRegistration, TaxRegistration{ID: idCode{Value: t.VATIdentifier, SchemeID: "VA"}})
	}
	return cp
}

func fromAddress(a model.Address) *Address {
	if a == (model.Address{}) {
		return nil
	}
	return &Address{
		PostcodeCode:       a.PostCode,
		LineOne:            a.Line1,
		LineTwo:            a.Line2,
		LineThree:          a.Line3,
		CityName:           a.City,
		CountryID:          a.CountryCode,
		CountrySubDivision: a.Subdivision,
	}
}

func fromContact(c *model.Contact) *Contact {
	if c == nil {
		return nil
	}
	cc := &Contact{PersonName: c.Point}
	if c.Phone != "" {
		cc.Telephone = &commNum{CompleteNumber: c.Phone}
	}
	if c.Email != "" {
		cc.Email = &commURI{URIID: c.Email}
	}
	return cc
}

func fromPaymentInstructions(inv *model.Invoice, st *Settlement) {
	pi := inv.PaymentInstructions
	if pi == nil {
		return
	}
	st.PaymentReference = pi.RemittanceInformation
	for _, ct := range pi.CreditTransfers {
		pm := PaymentMeans{TypeCode: pi.MeansTypeCode, Information: pi.MeansText}
		pm.PayeeAccount = &finAccount{IBANID: ct.AccountIdentifier, AccountName: ct.AccountName}
		if ct.ServiceProviderID != "" {
			pm.PayeeInstitution = &institution{BICID: ct.ServiceProviderID}
		}
		st.PaymentMeans = append(st.PaymentMeans, pm)
	}
	if pi.Card != nil {
		st.PaymentMeans = append(st.PaymentMeans, PaymentMeans{
			TypeCode:      pi.MeansTypeCode,
			Information:   pi.MeansText,
			FinancialCard: &Card{ID: pi.Card.PrimaryAccountNumber, CardholderName: pi.Card.HolderName},
		})
	}
	if dd := pi.DirectDebit; dd != nil {
		pm := PaymentMeans{TypeCode: pi.MeansTypeCode, Information: pi.MeansText}
		if dd.DebitedAccount != "" {
			pm.PayerAccount = &finAccount{IBANID: dd.DebitedAccount}
		}
		st.PaymentMeans = append(st.PaymentMeans, pm)
		st.CreditorReferenceID = dd.CreditorIdentifier
	}
	if len(st.PaymentMeans) == 0 && pi.MeansTypeCode != "" {
		st.PaymentMeans = append(st.PaymentMeans, PaymentMeans{TypeCode: pi.MeansTypeCode, Information: pi.MeansText})
	}
}

func fromPaymentTerms(inv *model.Invoice) *PaymentTerms {
	var mandate string
	if inv.PaymentInstructions != nil && inv.PaymentInstructions.DirectDebit != nil {
		mandate = inv.PaymentInstructions.DirectDebit.MandateReference
	}
	if inv.PaymentTerms == "" && inv.DueDate == "" && mandate == "" {
		return nil
	}
	return &PaymentTerms{
		Description:          inv.PaymentTerms,
		DueDateDateTime:      dateWrapOf(inv.DueDate),
		DirectDebitMandateID: mandate,
	}
}

func fromAllowanceCharge(a model.AllowanceCharge, charge bool) AllowanceCharge {
	ac := AllowanceCharge{
		ChargeIndicator:    indicator{Indicator: charge},
		CalculationPercent: string(a.Percentage),
		BasisAmount:        string(a.BaseAmount),
		ActualAmount:       string(a.Amount),
		ReasonCode:         a.ReasonCode,
		Reason:             a.Reason,
	}
	if a.VATCategoryCode != "" || a.VATRate != "" {
		ac.CategoryTradeTax = &CategoryTax{TypeCode: "VAT", CategoryCode: a.VATCategoryCode, RateApplicablePercent: string(a.VATRate)}
	}
	return ac
}

func fromTotals(t model.DocumentTotals, cur, acctCur string) Summation {
	s := Summation{
		LineTotalAmount:      string(t.LineNetTotal),
		ChargeTotalAmount:    string(t.ChargeTotal),
		AllowanceTotalAmount: string(t.AllowanceTotal),
		TaxBasisTotalAmount:  string(t.TaxBasisTotal),
		RoundingAmount:       string(t.RoundingAmount),
		GrandTotalAmount:     string(t.GrandTotal),
		TotalPrepaidAmount:   string(t.PaidAmount),
		DuePayableAmount:     string(t.DuePayableAmount),
	}
	s.TaxTotalAmount = append(s.TaxTotalAmount, taxAmount{Value: string(t.TaxTotal), Currency: cur})
	if t.TaxTotalAccountingCurrency != "" {
		s.TaxTotalAmount = append(s.TaxTotalAmount, taxAmount{Value: string(t.TaxTotalAccountingCurrency), Currency: acctCur})
	}
	return s
}

// fromLine maps a model line to a CII line. Sub-invoice lines (BG-DEX-01) are an
// XRechnung Extension defined for UBL only and have no CII representation, so
// l.SubLines is not emitted here.
func fromLine(l model.Line) LineItem {
	li := LineItem{}
	li.LineDocument = LineDocument{LineID: l.ID}
	if l.Note != "" {
		li.LineDocument.IncludedNote = &Note{Content: l.Note}
	}
	li.Product = fromItem(l.Item)
	li.Agreement = fromPrice(l.Price)
	if l.OrderLineReference != "" {
		li.Agreement.BuyerOrderReferencedDocument = &lineRef{LineID: l.OrderLineReference}
	}
	li.Delivery = LineDelivery{BilledQuantity: quantity{Value: string(l.Quantity.Value), Unit: l.Quantity.Unit}}
	li.Settlement = LineSettlement{
		ApplicableTradeTax: LineTax{
			TypeCode:              "VAT",
			ExemptionReason:       l.VAT.ExemptionReasonText,
			CategoryCode:          l.VAT.CategoryCode,
			ExemptionReasonCode:   l.VAT.ExemptionReasonCode,
			RateApplicablePercent: string(l.VAT.Rate),
		},
		BillingPeriod: fromPeriod(l.Period),
		Summation:     LineSummation{LineTotalAmount: string(l.NetAmount)},
	}
	if l.BuyerAccountingReference != "" {
		li.Settlement.AccountingAccount = &account{ID: l.BuyerAccountingReference}
	}
	if l.ObjectIdentifier != nil {
		li.Settlement.ObjectDocument = &AddlDoc{IssuerAssignedID: l.ObjectIdentifier.Value, TypeCode: "130", ReferenceTypeCode: l.ObjectIdentifier.Scheme}
	}
	for _, a := range l.Allowances {
		li.Settlement.AllowanceCharge = append(li.Settlement.AllowanceCharge, fromLineAllowance(a, false))
	}
	for _, ch := range l.Charges {
		li.Settlement.AllowanceCharge = append(li.Settlement.AllowanceCharge, fromLineAllowance(ch, true))
	}
	return li
}

func fromLineAllowance(a model.LineAllowanceCharge, charge bool) AllowanceCharge {
	return AllowanceCharge{
		ChargeIndicator:    indicator{Indicator: charge},
		CalculationPercent: string(a.Percentage),
		BasisAmount:        string(a.BaseAmount),
		ActualAmount:       string(a.Amount),
		ReasonCode:         a.ReasonCode,
		Reason:             a.Reason,
	}
}

func fromItem(it model.Item) Product {
	p := Product{
		Name:             it.Name,
		Description:      it.Description,
		SellerAssignedID: it.SellerIdentifier,
		BuyerAssignedID:  it.BuyerIdentifier,
	}
	if it.StandardIdentifier != nil {
		p.GlobalID = &idCode{Value: it.StandardIdentifier.Value, SchemeID: it.StandardIdentifier.Scheme}
	}
	if it.CountryOfOrigin != "" {
		p.OriginCountry = &country{ID: it.CountryOfOrigin}
	}
	for _, c := range it.Classifications {
		p.Classification = append(p.Classification, Classification{ClassCode: listCode{Value: c.Code, ListID: c.ListID, ListVersion: c.ListVersion}})
	}
	for _, a := range it.Attributes {
		p.Characteristic = append(p.Characteristic, Characteristic{Description: a.Name, Value: a.Value})
	}
	return p
}

func fromPrice(p model.Price) LineAgreement {
	la := LineAgreement{NetPrice: NetPrice{ChargeAmount: string(p.NetPrice)}}
	if p.BaseQuantity != nil {
		la.NetPrice.BasisQuantity = &quantity{Value: string(p.BaseQuantity.Value), Unit: p.BaseQuantity.Unit}
	}
	if p.GrossPrice != "" || p.Discount != "" {
		g := &GrossPrice{ChargeAmount: string(p.GrossPrice)}
		if p.Discount != "" {
			g.Discount = &PriceDiscount{ChargeIndicator: indicator{Indicator: false}, ActualAmount: string(p.Discount)}
		}
		la.GrossPrice = g
	}
	return la
}

func fromPeriod(p *model.Period) *period {
	if p == nil {
		return nil
	}
	return &period{StartDateTime: dateWrapOf(p.Start), EndDateTime: dateWrapOf(p.End)}
}

// dateWrapOf builds a udt:DateTimeString (format 102) wrapper, or nil if empty.
func dateWrapOf(d model.Date) *dateWrap {
	if d == "" {
		return nil
	}
	return &dateWrap{DateTimeString: dateStr{Format: "102", Value: toCIIDate(string(d))}}
}

func qdtDateWrapOf(d model.Date) *qdtDateWrap {
	if d == "" {
		return nil
	}
	return &qdtDateWrap{DateTimeString: dateStr{Format: "102", Value: toCIIDate(string(d))}}
}

// toCIIDate turns "2016-06-21" into "20160621".
func toCIIDate(s string) string { return strings.ReplaceAll(s, "-", "") }
