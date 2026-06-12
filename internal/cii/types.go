// Package cii converts between the syntax-neutral [model.Invoice] and the
// UN/CEFACT Cross Industry Invoice (CII) D16B syntax (the EN16931 / XRechnung
// subset).
//
// Like the ubl package, the XML structs use full namespace-URI tags so a value
// round-trips without losing namespace information. CII expresses dates as
// YYYYMMDD strings (format code 102); conversion to/from the model's YYYY-MM-DD
// is handled in the to/from-model code.
package cii

import "encoding/xml"

// CII namespace URIs (kept for reference; struct tags spell them out literally).
const (
	nsRSM = "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
	nsRAM = "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
	nsUDT = "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100"
	nsQDT = "urn:un:unece:uncefact:data:standard:QualifiedDataType:100"
)

// ---------------------------------------------------------------------------
// Leaf helper types
// ---------------------------------------------------------------------------

type idCode struct {
	Value    string `xml:",chardata"`
	SchemeID string `xml:"schemeID,attr,omitempty"`
}

type listCode struct {
	Value       string `xml:",chardata"`
	ListID      string `xml:"listID,attr,omitempty"`
	ListVersion string `xml:"listVersionID,attr,omitempty"`
}

type quantity struct {
	Value string `xml:",chardata"`
	Unit  string `xml:"unitCode,attr,omitempty"`
}

type taxAmount struct {
	Value    string `xml:",chardata"`
	Currency string `xml:"currencyID,attr,omitempty"`
}

type binary struct {
	Value    string `xml:",chardata"`
	MimeCode string `xml:"mimeCode,attr,omitempty"`
	Filename string `xml:"filename,attr,omitempty"`
}

// dateStr is the udt:DateTimeString leaf (format "102" = YYYYMMDD).
type dateStr struct {
	Format string `xml:"format,attr,omitempty"`
	Value  string `xml:",chardata"`
}

// dateWrap wraps a udt:DateTimeString inside elements such as IssueDateTime.
type dateWrap struct {
	DateTimeString dateStr `xml:"urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100 DateTimeString"`
}

// qdtDateWrap wraps a qdt:DateTimeString (used by FormattedIssueDateTime, BT-26).
type qdtDateWrap struct {
	DateTimeString dateStr `xml:"urn:un:unece:uncefact:data:standard:QualifiedDataType:100 DateTimeString"`
}

// dateOnlyWrap wraps a udt:DateString (used by TaxPointDate, BT-7).
type dateOnlyWrap struct {
	DateString dateStr `xml:"urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100 DateString"`
}

// indicator is the nested ChargeIndicator (udt:Indicator).
type indicator struct {
	Indicator bool `xml:"urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100 Indicator"`
}

// ---------------------------------------------------------------------------
// Document root
// ---------------------------------------------------------------------------

// Invoice is the CII CrossIndustryInvoice document root.
type Invoice struct {
	XMLName     xml.Name        `xml:"urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100 CrossIndustryInvoice"`
	Context     DocumentContext `xml:"urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100 ExchangedDocumentContext"`
	Document    ExchangedDoc    `xml:"urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100 ExchangedDocument"`
	Transaction Transaction     `xml:"urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100 SupplyChainTradeTransaction"`
}

// DocumentContext carries BT-23 (business process) and BT-24 (specification id).
type DocumentContext struct {
	BusinessProcess *paramID `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BusinessProcessSpecifiedDocumentContextParameter,omitempty"`
	Guideline       paramID  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 GuidelineSpecifiedDocumentContextParameter"`
}

type paramID struct {
	ID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID"`
}

// ExchangedDoc is the document header (BT-1/2/3, BG-1 notes).
type ExchangedDoc struct {
	ID            string    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID"`
	TypeCode      string    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TypeCode"`
	IssueDateTime *dateWrap `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IssueDateTime,omitempty"`
	IncludedNote  []Note    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IncludedNote,omitempty"`
}

// Note is a document or line note (BG-1: BT-22 content, BT-21 subject code).
type Note struct {
	Content     string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Content,omitempty"`
	SubjectCode string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SubjectCode,omitempty"`
}

// Transaction holds the lines and the three header trade sections.
type Transaction struct {
	Lines      []LineItem `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IncludedSupplyChainTradeLineItem"`
	Agreement  Agreement  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ApplicableHeaderTradeAgreement"`
	Delivery   Delivery   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ApplicableHeaderTradeDelivery"`
	Settlement Settlement `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ApplicableHeaderTradeSettlement"`
}

// ---------------------------------------------------------------------------
// Lines (BG-25)
// ---------------------------------------------------------------------------

type LineItem struct {
	LineDocument LineDocument   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 AssociatedDocumentLineDocument"`
	Product      Product        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTradeProduct"`
	Agreement    LineAgreement  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedLineTradeAgreement"`
	Delivery     LineDelivery   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedLineTradeDelivery"`
	Settlement   LineSettlement `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedLineTradeSettlement"`
}

type LineDocument struct {
	LineID       string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 LineID"`
	IncludedNote *Note  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IncludedNote,omitempty"`
}

type Product struct {
	GlobalID         *idCode          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 GlobalID,omitempty"`
	SellerAssignedID string           `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SellerAssignedID,omitempty"`
	BuyerAssignedID  string           `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BuyerAssignedID,omitempty"`
	Name             string           `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Name,omitempty"`
	Description      string           `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Description,omitempty"`
	Characteristic   []Characteristic `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ApplicableProductCharacteristic,omitempty"`
	Classification   []Classification `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DesignatedProductClassification,omitempty"`
	OriginCountry    *country         `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 OriginTradeCountry,omitempty"`
}

type Characteristic struct {
	Description string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Description,omitempty"`
	Value       string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Value,omitempty"`
}

type Classification struct {
	ClassCode listCode `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ClassCode"`
}

type country struct {
	ID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID,omitempty"`
}

type LineAgreement struct {
	BuyerOrderReferencedDocument *lineRef    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BuyerOrderReferencedDocument,omitempty"`
	GrossPrice                   *GrossPrice `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 GrossPriceProductTradePrice,omitempty"`
	NetPrice                     NetPrice    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 NetPriceProductTradePrice"`
}

type lineRef struct {
	LineID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 LineID,omitempty"`
}

type GrossPrice struct {
	ChargeAmount  string         `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ChargeAmount"`
	BasisQuantity *quantity      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BasisQuantity,omitempty"`
	Discount      *PriceDiscount `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 AppliedTradeAllowanceCharge,omitempty"`
}

type PriceDiscount struct {
	ChargeIndicator indicator `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ChargeIndicator"`
	ActualAmount    string    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ActualAmount"`
}

type NetPrice struct {
	ChargeAmount  string    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ChargeAmount"`
	BasisQuantity *quantity `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BasisQuantity,omitempty"`
}

type LineDelivery struct {
	BilledQuantity quantity `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BilledQuantity"`
}

type LineSettlement struct {
	ApplicableTradeTax LineTax           `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ApplicableTradeTax"`
	BillingPeriod      *period           `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BillingSpecifiedPeriod,omitempty"`
	AllowanceCharge    []AllowanceCharge `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTradeAllowanceCharge,omitempty"`
	Summation          LineSummation     `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTradeSettlementLineMonetarySummation"`
	ObjectDocument     *AddlDoc          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 AdditionalReferencedDocument,omitempty"`
	AccountingAccount  *account          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ReceivableSpecifiedTradeAccountingAccount,omitempty"`
}

type LineTax struct {
	TypeCode              string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TypeCode,omitempty"`
	ExemptionReason       string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ExemptionReason,omitempty"`
	CategoryCode          string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CategoryCode,omitempty"`
	ExemptionReasonCode   string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ExemptionReasonCode,omitempty"`
	RateApplicablePercent string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 RateApplicablePercent,omitempty"`
}

type LineSummation struct {
	LineTotalAmount string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 LineTotalAmount"`
}

type account struct {
	ID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID,omitempty"`
}

// ---------------------------------------------------------------------------
// Header trade agreement (BG-4, BG-7, references)
// ---------------------------------------------------------------------------

type Agreement struct {
	BuyerReference             string    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BuyerReference,omitempty"`
	SellerTradeParty           Party     `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SellerTradeParty"`
	BuyerTradeParty            Party     `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BuyerTradeParty"`
	TaxRepresentativeParty     *Party    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SellerTaxRepresentativeTradeParty,omitempty"`
	SellerOrderReferencedDoc   *docRef   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SellerOrderReferencedDocument,omitempty"`
	BuyerOrderReferencedDoc    *docRef   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BuyerOrderReferencedDocument,omitempty"`
	ContractReferencedDocument *docRef   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ContractReferencedDocument,omitempty"`
	AdditionalReferencedDoc    []AddlDoc `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 AdditionalReferencedDocument,omitempty"`
	ProcuringProject           *project  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedProcuringProject,omitempty"`
}

type docRef struct {
	IssuerAssignedID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IssuerAssignedID,omitempty"`
}

type project struct {
	ID   string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID,omitempty"`
	Name string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Name,omitempty"`
}

// AddlDoc is an AdditionalReferencedDocument: BT-18 (TypeCode 130), BT-17
// (TypeCode 50), BT-128 (line, TypeCode 130), or BG-24 supporting document.
type AddlDoc struct {
	IssuerAssignedID  string  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IssuerAssignedID,omitempty"`
	URIID             string  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 URIID,omitempty"`
	TypeCode          string  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TypeCode,omitempty"`
	Name              string  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Name,omitempty"`
	AttachmentBinary  *binary `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 AttachmentBinaryObject,omitempty"`
	ReferenceTypeCode string  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ReferenceTypeCode,omitempty"`
}

// ---------------------------------------------------------------------------
// Parties
// ---------------------------------------------------------------------------

type Party struct {
	ID              []string          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID,omitempty"`
	GlobalID        []idCode          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 GlobalID,omitempty"`
	Name            string            `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Name,omitempty"`
	Description     string            `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Description,omitempty"`
	LegalOrg        *LegalOrg         `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedLegalOrganization,omitempty"`
	Contact         *Contact          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DefinedTradeContact,omitempty"`
	PostalAddress   *Address          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PostalTradeAddress,omitempty"`
	URIComm         *uriComm          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 URIUniversalCommunication,omitempty"`
	TaxRegistration []TaxRegistration `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTaxRegistration,omitempty"`
}

type LegalOrg struct {
	ID                  *idCode `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID,omitempty"`
	TradingBusinessName string  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TradingBusinessName,omitempty"`
}

type Contact struct {
	PersonName     string   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PersonName,omitempty"`
	DepartmentName string   `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DepartmentName,omitempty"`
	Telephone      *commNum `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TelephoneUniversalCommunication,omitempty"`
	Email          *commURI `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 EmailURIUniversalCommunication,omitempty"`
}

type commNum struct {
	CompleteNumber string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CompleteNumber,omitempty"`
}

type commURI struct {
	URIID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 URIID,omitempty"`
}

type uriComm struct {
	URIID idCode `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 URIID"`
}

type Address struct {
	PostcodeCode       string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PostcodeCode,omitempty"`
	LineOne            string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 LineOne,omitempty"`
	LineTwo            string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 LineTwo,omitempty"`
	LineThree          string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 LineThree,omitempty"`
	CityName           string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CityName,omitempty"`
	CountryID          string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CountryID,omitempty"`
	CountrySubDivision string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CountrySubDivisionName,omitempty"`
}

type TaxRegistration struct {
	ID idCode `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID"`
}

// ---------------------------------------------------------------------------
// Header trade delivery (BG-13/14/15)
// ---------------------------------------------------------------------------

type Delivery struct {
	ShipToTradeParty   *Party         `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ShipToTradeParty,omitempty"`
	ActualDelivery     *deliveryEvent `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ActualDeliverySupplyChainEvent,omitempty"`
	DespatchAdviceRef  *docRef        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DespatchAdviceReferencedDocument,omitempty"`
	ReceivingAdviceRef *docRef        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ReceivingAdviceReferencedDocument,omitempty"`
}

type deliveryEvent struct {
	OccurrenceDateTime *dateWrap `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 OccurrenceDateTime,omitempty"`
}

// ---------------------------------------------------------------------------
// Header trade settlement (payment, tax, totals)
// ---------------------------------------------------------------------------

type Settlement struct {
	CreditorReferenceID  string            `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CreditorReferenceID,omitempty"`
	PaymentReference     string            `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PaymentReference,omitempty"`
	TaxCurrencyCode      string            `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TaxCurrencyCode,omitempty"`
	InvoiceCurrencyCode  string            `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 InvoiceCurrencyCode,omitempty"`
	PayeeTradeParty      *Party            `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PayeeTradeParty,omitempty"`
	PaymentMeans         []PaymentMeans    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTradeSettlementPaymentMeans,omitempty"`
	ApplicableTradeTax   []TradeTax        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ApplicableTradeTax,omitempty"`
	BillingPeriod        *period           `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BillingSpecifiedPeriod,omitempty"`
	AllowanceCharge      []AllowanceCharge `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTradeAllowanceCharge,omitempty"`
	PaymentTerms         *PaymentTerms     `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTradePaymentTerms,omitempty"`
	Summation            Summation         `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 SpecifiedTradeSettlementHeaderMonetarySummation"`
	InvoiceReferencedDoc []InvoiceRef      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 InvoiceReferencedDocument,omitempty"`
	AccountingAccount    *account          `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ReceivableSpecifiedTradeAccountingAccount,omitempty"`
}

type period struct {
	StartDateTime *dateWrap `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 StartDateTime,omitempty"`
	EndDateTime   *dateWrap `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 EndDateTime,omitempty"`
}

type PaymentMeans struct {
	TypeCode         string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TypeCode,omitempty"`
	Information      string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Information,omitempty"`
	FinancialCard    *Card        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ApplicableTradeSettlementFinancialCard,omitempty"`
	PayerAccount     *finAccount  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PayerPartyDebtorFinancialAccount,omitempty"`
	PayeeAccount     *finAccount  `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PayeePartyCreditorFinancialAccount,omitempty"`
	PayeeInstitution *institution `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 PayeeSpecifiedCreditorFinancialInstitution,omitempty"`
}

type Card struct {
	ID             string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ID,omitempty"`
	CardholderName string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CardholderName,omitempty"`
}

type finAccount struct {
	IBANID        string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IBANID,omitempty"`
	AccountName   string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 AccountName,omitempty"`
	ProprietaryID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ProprietaryID,omitempty"`
}

type institution struct {
	BICID string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BICID,omitempty"`
}

type TradeTax struct {
	CalculatedAmount      string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CalculatedAmount"`
	TypeCode              string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TypeCode,omitempty"`
	ExemptionReason       string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ExemptionReason,omitempty"`
	BasisAmount           string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BasisAmount,omitempty"`
	CategoryCode          string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CategoryCode,omitempty"`
	ExemptionReasonCode   string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ExemptionReasonCode,omitempty"`
	TaxPointDate          *dateOnlyWrap `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TaxPointDate,omitempty"`
	DueDateTypeCode       string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DueDateTypeCode,omitempty"`
	RateApplicablePercent string        `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 RateApplicablePercent,omitempty"`
}

// AllowanceCharge is a document-level (BG-20/21) or line-level (BG-27/28)
// allowance/charge. CategoryTradeTax is only used at document level.
type AllowanceCharge struct {
	ChargeIndicator    indicator    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ChargeIndicator"`
	CalculationPercent string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CalculationPercent,omitempty"`
	BasisAmount        string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 BasisAmount,omitempty"`
	ActualAmount       string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ActualAmount"`
	ReasonCode         string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ReasonCode,omitempty"`
	Reason             string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Reason,omitempty"`
	CategoryTradeTax   *CategoryTax `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CategoryTradeTax,omitempty"`
}

type CategoryTax struct {
	TypeCode              string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TypeCode,omitempty"`
	CategoryCode          string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 CategoryCode,omitempty"`
	RateApplicablePercent string `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 RateApplicablePercent,omitempty"`
}

type PaymentTerms struct {
	Description          string    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 Description,omitempty"`
	DueDateDateTime      *dateWrap `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DueDateDateTime,omitempty"`
	DirectDebitMandateID string    `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DirectDebitMandateID,omitempty"`
}

type Summation struct {
	LineTotalAmount      string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 LineTotalAmount"`
	ChargeTotalAmount    string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 ChargeTotalAmount,omitempty"`
	AllowanceTotalAmount string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 AllowanceTotalAmount,omitempty"`
	TaxBasisTotalAmount  string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TaxBasisTotalAmount"`
	TaxTotalAmount       []taxAmount `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TaxTotalAmount,omitempty"`
	RoundingAmount       string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 RoundingAmount,omitempty"`
	GrandTotalAmount     string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 GrandTotalAmount"`
	TotalPrepaidAmount   string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 TotalPrepaidAmount,omitempty"`
	DuePayableAmount     string      `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 DuePayableAmount"`
}

type InvoiceRef struct {
	IssuerAssignedID       string       `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 IssuerAssignedID,omitempty"`
	FormattedIssueDateTime *qdtDateWrap `xml:"urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100 FormattedIssueDateTime,omitempty"`
}
