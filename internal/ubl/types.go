// Package ubl converts between the syntax-neutral [model.Invoice] and the OASIS
// UBL 2.1 Invoice syntax (the EN16931 / XRechnung subset).
//
// The XML structs use full namespace-URI tags so that a value round-trips
// (unmarshal -> marshal -> unmarshal) without losing namespace information.
// Marshal output is namespace-correct (each element carries its xmlns) though
// not prefixed; readability of the serialized form is a later concern.
package ubl

import "encoding/xml"

// UBL namespace URIs. Kept as constants for reference; struct tags must spell
// the URI out literally.
const (
	nsInvoice    = "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
	nsCreditNote = "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2"
	nsCAC        = "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
	nsCBC        = "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
)

// ---------------------------------------------------------------------------
// Reusable leaf types
// ---------------------------------------------------------------------------

// amount is a monetary value with its currency attribute.
type amount struct {
	Value    string `xml:",chardata"`
	Currency string `xml:"currencyID,attr,omitempty"`
}

// idCode is an identifier/code value with an optional scheme attribute.
type idCode struct {
	Value    string `xml:",chardata"`
	SchemeID string `xml:"schemeID,attr,omitempty"`
}

// listCode is a classification code with list attributes (BT-158).
type listCode struct {
	Value       string `xml:",chardata"`
	ListID      string `xml:"listID,attr,omitempty"`
	ListVersion string `xml:"listVersionID,attr,omitempty"`
}

// quantity is a numeric value with a unit-of-measure code.
type quantity struct {
	Value string `xml:",chardata"`
	Unit  string `xml:"unitCode,attr,omitempty"`
}

// binary is an embedded base64 document with its mime code and filename.
type binary struct {
	Value    string `xml:",chardata"`
	MimeCode string `xml:"mimeCode,attr,omitempty"`
	Filename string `xml:"filename,attr,omitempty"`
}

// typedCode is a code value with an optional name attribute (e.g. payment means
// code BT-81 with text BT-82).
type typedCode struct {
	Value string `xml:",chardata"`
	Name  string `xml:"name,attr,omitempty"`
}

// ---------------------------------------------------------------------------
// Document root
// ---------------------------------------------------------------------------

// Invoice is the UBL Invoice document root.
type Invoice struct {
	// XMLName is left untagged so the same struct unmarshals both the UBL Invoice
	// root and the CreditNote root; Marshal sets it explicitly.
	XMLName xml.Name

	CustomizationID      string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CustomizationID,omitempty"`
	ProfileID            string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ProfileID,omitempty"`
	ID                   string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID"`
	IssueDate            string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 IssueDate,omitempty"`
	DueDate              string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 DueDate,omitempty"`
	InvoiceTypeCode      string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 InvoiceTypeCode,omitempty"`
	CreditNoteTypeCode   string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CreditNoteTypeCode,omitempty"`
	Note                 []string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Note,omitempty"`
	TaxPointDate         string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxPointDate,omitempty"`
	DocumentCurrencyCode string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 DocumentCurrencyCode,omitempty"`
	TaxCurrencyCode      string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxCurrencyCode,omitempty"`
	AccountingCost       string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 AccountingCost,omitempty"`
	BuyerReference       string   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 BuyerReference,omitempty"`

	InvoicePeriod               *Period            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 InvoicePeriod,omitempty"`
	OrderReference              *OrderReference    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 OrderReference,omitempty"`
	BillingReference            []BillingReference `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 BillingReference,omitempty"`
	DespatchDocumentReference   *DocRef            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 DespatchDocumentReference,omitempty"`
	ReceiptDocumentReference    *DocRef            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 ReceiptDocumentReference,omitempty"`
	OriginatorDocumentReference *DocRef            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 OriginatorDocumentReference,omitempty"`
	ContractDocumentReference   *DocRef            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 ContractDocumentReference,omitempty"`
	ProjectReference            *DocRef            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 ProjectReference,omitempty"`
	AdditionalDocumentReference []AdditionalDocRef `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AdditionalDocumentReference,omitempty"`

	AccountingSupplierParty SupplierParty     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AccountingSupplierParty"`
	AccountingCustomerParty CustomerParty     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AccountingCustomerParty"`
	PayeeParty              *Party            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PayeeParty,omitempty"`
	TaxRepresentativeParty  *Party            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 TaxRepresentativeParty,omitempty"`
	Delivery                *Delivery         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Delivery,omitempty"`
	PaymentMeans            []PaymentMeans    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PaymentMeans,omitempty"`
	PaymentTerms            *PaymentTerms     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PaymentTerms,omitempty"`
	PrepaidPayment          []PrepaidPayment  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PrepaidPayment,omitempty"`
	AllowanceCharge         []AllowanceCharge `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AllowanceCharge,omitempty"`
	TaxTotal                []TaxTotal        `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 TaxTotal,omitempty"`
	LegalMonetaryTotal      MonetaryTotal     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 LegalMonetaryTotal"`
	InvoiceLine             []InvoiceLine     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 InvoiceLine,omitempty"`
	CreditNoteLine          []InvoiceLine     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 CreditNoteLine,omitempty"`
}

// Period is a date range (cac:InvoicePeriod).
type Period struct {
	StartDate string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 StartDate,omitempty"`
	EndDate   string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 EndDate,omitempty"`
}

// OrderReference carries BT-13 (ID) and BT-14 (SalesOrderID).
type OrderReference struct {
	ID           string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
	SalesOrderID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 SalesOrderID,omitempty"`
}

// BillingReference references a preceding invoice (BG-3).
type BillingReference struct {
	InvoiceDocumentReference DocRefWithDate `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 InvoiceDocumentReference"`
}

// DocRef is a simple document reference with just an ID.
type DocRef struct {
	ID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
}

// DocRefWithDate is a document reference with an issue date (BT-25/26).
type DocRefWithDate struct {
	ID        string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
	IssueDate string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 IssueDate,omitempty"`
}

// AdditionalDocRef is a supporting document (BG-24) or, when DocumentTypeCode is
// "130", the invoiced object identifier (BT-18).
type AdditionalDocRef struct {
	ID                  idCode      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID"`
	DocumentTypeCode    string      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 DocumentTypeCode,omitempty"`
	DocumentDescription string      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 DocumentDescription,omitempty"`
	Attachment          *Attachment `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Attachment,omitempty"`
}

// Attachment holds an embedded or externally referenced document (BT-124/125).
type Attachment struct {
	EmbeddedDocumentBinaryObject *binary `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 EmbeddedDocumentBinaryObject,omitempty"`
	ExternalReference            *ExtRef `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 ExternalReference,omitempty"`
}

// ExtRef is an external document location (BT-124).
type ExtRef struct {
	URI string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 URI,omitempty"`
}

// ---------------------------------------------------------------------------
// Parties
// ---------------------------------------------------------------------------

// SupplierParty wraps the Seller party (BG-4).
type SupplierParty struct {
	Party Party `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Party"`
}

// CustomerParty wraps the Buyer party (BG-7).
type CustomerParty struct {
	Party Party `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Party"`
}

// Party is a UBL party (Seller, Buyer, Payee, Tax representative, Deliver to).
type Party struct {
	EndpointID          *idCode          `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 EndpointID,omitempty"`
	PartyIdentification []PartyID        `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PartyIdentification,omitempty"`
	PartyName           *PartyName       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PartyName,omitempty"`
	PostalAddress       *Address         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PostalAddress,omitempty"`
	PartyTaxScheme      []PartyTaxScheme `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PartyTaxScheme,omitempty"`
	PartyLegalEntity    *LegalEntity     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PartyLegalEntity,omitempty"`
	Contact             *Contact         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Contact,omitempty"`
}

// PartyID is a party identification (BT-29/46/60).
type PartyID struct {
	ID idCode `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID"`
}

// PartyName is a party trading name (BT-28/45).
type PartyName struct {
	Name string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Name,omitempty"`
}

// Address is a UBL postal address (BG-5/8/12/15).
type Address struct {
	StreetName           string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 StreetName,omitempty"`
	AdditionalStreetName string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 AdditionalStreetName,omitempty"`
	CityName             string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CityName,omitempty"`
	PostalZone           string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PostalZone,omitempty"`
	CountrySubentity     string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CountrySubentity,omitempty"`
	AddressLine          *AddressLine `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AddressLine,omitempty"`
	Country              Country      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Country"`
}

// AddressLine is the third address line (BT-162/163/164/165).
type AddressLine struct {
	Line string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Line,omitempty"`
}

// Country holds a country identification code.
type Country struct {
	IdentificationCode string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 IdentificationCode,omitempty"`
}

// PartyTaxScheme carries a VAT or tax-registration identifier (BT-31/32/48/63).
type PartyTaxScheme struct {
	CompanyID string    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CompanyID,omitempty"`
	TaxScheme TaxScheme `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 TaxScheme"`
}

// LegalEntity carries the registered name and registration id (BT-27/30/33).
type LegalEntity struct {
	RegistrationName string  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 RegistrationName,omitempty"`
	CompanyID        *idCode `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CompanyID,omitempty"`
	CompanyLegalForm string  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CompanyLegalForm,omitempty"`
}

// Contact is a party contact (BG-6/9).
type Contact struct {
	Name           string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Name,omitempty"`
	Telephone      string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Telephone,omitempty"`
	ElectronicMail string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ElectronicMail,omitempty"`
}

// TaxScheme identifies the tax scheme (almost always "VAT").
type TaxScheme struct {
	ID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
}

// ---------------------------------------------------------------------------
// Delivery
// ---------------------------------------------------------------------------

// Delivery holds delivery information (BG-13/14/15).
type Delivery struct {
	ActualDeliveryDate string            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ActualDeliveryDate,omitempty"`
	DeliveryLocation   *DeliveryLocation `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 DeliveryLocation,omitempty"`
	DeliveryParty      *Party            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 DeliveryParty,omitempty"`
}

// DeliveryLocation holds the deliver-to location id and address.
type DeliveryLocation struct {
	ID      *idCode  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
	Address *Address `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Address,omitempty"`
}

// ---------------------------------------------------------------------------
// Payment
// ---------------------------------------------------------------------------

// PaymentMeans describes how payment is made (BG-16/17/18/19).
type PaymentMeans struct {
	PaymentMeansCode      typedCode         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PaymentMeansCode"`
	PaymentID             string            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PaymentID,omitempty"`
	CardAccount           *CardAccount      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 CardAccount,omitempty"`
	PayeeFinancialAccount *FinancialAccount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PayeeFinancialAccount,omitempty"`
	PaymentMandate        *PaymentMandate   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PaymentMandate,omitempty"`
}

// CardAccount is payment-card information (BG-18).
type CardAccount struct {
	PrimaryAccountNumberID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PrimaryAccountNumberID,omitempty"`
	NetworkID              string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 NetworkID,omitempty"`
	HolderName             string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 HolderName,omitempty"`
}

// FinancialAccount is a credit-transfer account (BG-17).
type FinancialAccount struct {
	ID                         string                      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
	Name                       string                      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Name,omitempty"`
	FinancialInstitutionBranch *FinancialInstitutionBranch `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 FinancialInstitutionBranch,omitempty"`
}

// FinancialInstitutionBranch carries the BIC (BT-86).
type FinancialInstitutionBranch struct {
	ID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
}

// PaymentMandate is direct-debit mandate information (BG-19).
type PaymentMandate struct {
	ID                    string            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
	PayerFinancialAccount *FinancialAccount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 PayerFinancialAccount,omitempty"`
}

// PaymentTerms carries the payment terms text (BT-20).
type PaymentTerms struct {
	Note string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Note,omitempty"`
}

// PrepaidPayment is a third-party payment (BG-DEX-09, XRechnung Extension).
type PrepaidPayment struct {
	ID            string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
	PaidAmount    amount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PaidAmount,omitempty"`
	InstructionID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 InstructionID,omitempty"`
}

// ---------------------------------------------------------------------------
// Allowance / charge, tax, totals
// ---------------------------------------------------------------------------

// AllowanceCharge is a document- or line-level allowance/charge (BG-20/21/27/28).
type AllowanceCharge struct {
	ChargeIndicator           bool         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ChargeIndicator"`
	AllowanceChargeReasonCode string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 AllowanceChargeReasonCode,omitempty"`
	AllowanceChargeReason     string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 AllowanceChargeReason,omitempty"`
	MultiplierFactorNumeric   string       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 MultiplierFactorNumeric,omitempty"`
	Amount                    amount       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Amount"`
	BaseAmount                *amount      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 BaseAmount,omitempty"`
	TaxCategory               *TaxCategory `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 TaxCategory,omitempty"`
}

// TaxTotal aggregates tax (BT-110) and its category breakdown (BG-23).
type TaxTotal struct {
	TaxAmount   amount        `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxAmount"`
	TaxSubtotal []TaxSubtotal `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 TaxSubtotal,omitempty"`
}

// TaxSubtotal is one VAT category subtotal (BG-23).
type TaxSubtotal struct {
	TaxableAmount amount      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxableAmount"`
	TaxAmount     amount      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxAmount"`
	TaxCategory   TaxCategory `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 TaxCategory"`
}

// TaxCategory is a VAT category (code, rate, exemption reason).
type TaxCategory struct {
	ID                     string    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
	Percent                string    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Percent,omitempty"`
	TaxExemptionReasonCode string    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxExemptionReasonCode,omitempty"`
	TaxExemptionReason     string    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxExemptionReason,omitempty"`
	TaxScheme              TaxScheme `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 TaxScheme"`
}

// MonetaryTotal holds the document totals (BG-22).
type MonetaryTotal struct {
	LineExtensionAmount   amount  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 LineExtensionAmount"`
	TaxExclusiveAmount    amount  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxExclusiveAmount"`
	TaxInclusiveAmount    amount  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 TaxInclusiveAmount"`
	AllowanceTotalAmount  *amount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 AllowanceTotalAmount,omitempty"`
	ChargeTotalAmount     *amount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ChargeTotalAmount,omitempty"`
	PrepaidAmount         *amount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PrepaidAmount,omitempty"`
	PayableRoundingAmount *amount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PayableRoundingAmount,omitempty"`
	PayableAmount         amount  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PayableAmount"`
}

// ---------------------------------------------------------------------------
// Invoice line
// ---------------------------------------------------------------------------

// InvoiceLine is a single invoice line (BG-25).
type InvoiceLine struct {
	ID                  string            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID"`
	Note                string            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Note,omitempty"`
	InvoicedQuantity    *quantity         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 InvoicedQuantity,omitempty"`
	CreditedQuantity    *quantity         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 CreditedQuantity,omitempty"`
	LineExtensionAmount amount            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 LineExtensionAmount"`
	AccountingCost      string            `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 AccountingCost,omitempty"`
	InvoicePeriod       *Period           `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 InvoicePeriod,omitempty"`
	OrderLineReference  *OrderLineRef     `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 OrderLineReference,omitempty"`
	DocumentReference   *LineDocRef       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 DocumentReference,omitempty"`
	AllowanceCharge     []AllowanceCharge `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AllowanceCharge,omitempty"`
	Item                Item              `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Item"`
	Price               Price             `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 Price"`
	// SubInvoiceLine carries nested sub-invoice lines (XRechnung Extension, BG-DEX-01).
	SubInvoiceLine []InvoiceLine `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 SubInvoiceLine,omitempty"`
}

// OrderLineRef references a purchase order line (BT-132).
type OrderLineRef struct {
	LineID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 LineID,omitempty"`
}

// LineDocRef is the invoice line object identifier (BT-128). UBL binds it to a
// DocumentReference whose DocumentTypeCode is the constant "130".
type LineDocRef struct {
	ID               idCode `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID"`
	DocumentTypeCode string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 DocumentTypeCode,omitempty"`
}

// Item is the line item information (BG-31).
type Item struct {
	Description                string           `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Description,omitempty"`
	Name                       string           `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Name,omitempty"`
	BuyersItemIdentification   *ItemID          `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 BuyersItemIdentification,omitempty"`
	SellersItemIdentification  *ItemID          `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 SellersItemIdentification,omitempty"`
	StandardItemIdentification *StandardItemID  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 StandardItemIdentification,omitempty"`
	OriginCountry              *Country         `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 OriginCountry,omitempty"`
	CommodityClassification    []CommodityClass `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 CommodityClassification,omitempty"`
	ClassifiedTaxCategory      TaxCategory      `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 ClassifiedTaxCategory"`
	AdditionalItemProperty     []ItemProperty   `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AdditionalItemProperty,omitempty"`
}

// ItemID is a seller/buyer item identifier (BT-155/156).
type ItemID struct {
	ID string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID,omitempty"`
}

// StandardItemID is a standard item identifier with scheme (BT-157).
type StandardItemID struct {
	ID idCode `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ID"`
}

// CommodityClass is an item classification (BT-158).
type CommodityClass struct {
	ItemClassificationCode listCode `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ItemClassificationCode"`
}

// ItemProperty is an item attribute name/value (BG-32).
type ItemProperty struct {
	Name  string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Name,omitempty"`
	Value string `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Value,omitempty"`
}

// Price is the line price details (BG-29).
type Price struct {
	PriceAmount     amount          `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 PriceAmount"`
	BaseQuantity    *quantity       `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 BaseQuantity,omitempty"`
	AllowanceCharge *PriceAllowance `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2 AllowanceCharge,omitempty"`
}

// PriceAllowance carries the price discount (BT-147) and gross price (BT-148).
type PriceAllowance struct {
	ChargeIndicator bool    `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 ChargeIndicator"`
	Amount          amount  `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 Amount"`
	BaseAmount      *amount `xml:"urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2 BaseAmount,omitempty"`
}
