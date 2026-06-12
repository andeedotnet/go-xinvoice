package model

// Invoice is the syntax-neutral EN16931 / XRechnung invoice. Field comments give
// the Business Term (BT-*) or Business Group (BG-*) each value represents.
//
// Optional single groups are pointers; repeatable groups are slices. Optional
// scalar values rely on `omitempty` (empty string / empty Decimal / empty Date).
type Invoice struct {
	Number                  string `json:"number"`                            // BT-1  Invoice number
	IssueDate               Date   `json:"issueDate"`                         // BT-2  Invoice issue date
	TypeCode                Code   `json:"typeCode"`                          // BT-3  Invoice type code
	CurrencyCode            Code   `json:"currencyCode"`                      // BT-5  Invoice currency code
	VATAccountingCurrency   Code   `json:"vatAccountingCurrency,omitempty"`   // BT-6  VAT accounting currency code
	TaxPointDate            Date   `json:"taxPointDate,omitempty"`            // BT-7  Value added tax point date
	TaxPointDateCode        Code   `json:"taxPointDateCode,omitempty"`        // BT-8  Value added tax point date code
	DueDate                 Date   `json:"dueDate,omitempty"`                 // BT-9  Payment due date
	BuyerReference          string `json:"buyerReference,omitempty"`          // BT-10 Buyer reference (Leitweg-ID)
	BusinessProcessType     string `json:"businessProcessType,omitempty"`     // BT-23 Business process type (BG-2)
	SpecificationIdentifier string `json:"specificationIdentifier,omitempty"` // BT-24 Specification identifier (BG-2)

	// Document-level references (BT-11..BT-17)
	ProjectReference         string `json:"projectReference,omitempty"`         // BT-11
	ContractReference        string `json:"contractReference,omitempty"`        // BT-12
	PurchaseOrderReference   string `json:"purchaseOrderReference,omitempty"`   // BT-13
	SalesOrderReference      string `json:"salesOrderReference,omitempty"`      // BT-14
	ReceivingAdviceReference string `json:"receivingAdviceReference,omitempty"` // BT-15
	DespatchAdviceReference  string `json:"despatchAdviceReference,omitempty"`  // BT-16
	TenderOrLotReference     string `json:"tenderOrLotReference,omitempty"`     // BT-17

	InvoicedObject           *ID    `json:"invoicedObject,omitempty"`           // BT-18 (+ BT-18-1 scheme)
	BuyerAccountingReference string `json:"buyerAccountingReference,omitempty"` // BT-19
	PaymentTerms             string `json:"paymentTerms,omitempty"`             // BT-20

	Notes             []Note                      `json:"notes,omitempty"`             // BG-1  (0..n)
	PrecedingInvoices []PrecedingInvoiceReference `json:"precedingInvoices,omitempty"` // BG-3  (0..n)

	Seller            Party              `json:"seller"`                      // BG-4 (+ BG-5, BG-6)
	Buyer             Party              `json:"buyer"`                       // BG-7 (+ BG-8, BG-9)
	Payee             *Payee             `json:"payee,omitempty"`             // BG-10
	TaxRepresentative *TaxRepresentative `json:"taxRepresentative,omitempty"` // BG-11 (+ BG-12)
	Delivery          *Delivery          `json:"delivery,omitempty"`          // BG-13 (+ BG-14, BG-15)

	PaymentInstructions *PaymentInstructions `json:"paymentInstructions,omitempty"` // BG-16 (+ BG-17..19)

	Allowances []AllowanceCharge `json:"allowances,omitempty"` // BG-20 (0..n)
	Charges    []AllowanceCharge `json:"charges,omitempty"`    // BG-21 (0..n)

	Totals             DocumentTotals       `json:"totals"`                       // BG-22
	VATBreakdown       []VATBreakdown       `json:"vatBreakdown"`                 // BG-23 (1..n)
	Documents          []SupportingDocument `json:"documents,omitempty"`          // BG-24 (0..n)
	ThirdPartyPayments []ThirdPartyPayment  `json:"thirdPartyPayments,omitempty"` // BG-DEX-09 (XRechnung Extension)
	Lines              []Line               `json:"lines"`                        // BG-25 (1..n)
}

// ThirdPartyPayment is a payment made by a third party. BG-DEX-09 (XRechnung
// Extension). In UBL it maps to cac:PrepaidPayment.
type ThirdPartyPayment struct {
	Type        string `json:"type,omitempty"`        // BT-DEX-001 Third party payment type
	Amount      Amount `json:"amount"`                // BT-DEX-002 Third party payment amount
	Description string `json:"description,omitempty"` // BT-DEX-003 Third party payment description
}

// Note is a free-text invoice note. BG-1.
type Note struct {
	SubjectCode Code   `json:"subjectCode,omitempty"` // BT-21 Invoice note subject code
	Text        string `json:"text"`                  // BT-22 Invoice note
}

// PrecedingInvoiceReference points at an earlier invoice. BG-3.
type PrecedingInvoiceReference struct {
	Reference string `json:"reference"`           // BT-25 Preceding Invoice reference
	IssueDate Date   `json:"issueDate,omitempty"` // BT-26 Preceding Invoice issue date
}

// Party represents the Seller (BG-4) or Buyer (BG-7). It carries the superset of
// both; Seller-only terms (TaxRegistrationID, AdditionalLegalInfo) stay empty for
// the Buyer.
type Party struct {
	Name                string   `json:"name"`                          // BT-27 / BT-44
	TradingName         string   `json:"tradingName,omitempty"`         // BT-28 / BT-45
	Identifiers         []ID     `json:"identifiers,omitempty"`         // BT-29(+1) / BT-46(+1)
	LegalRegistrationID *ID      `json:"legalRegistrationId,omitempty"` // BT-30(+1) / BT-47(+1)
	VATIdentifier       string   `json:"vatIdentifier,omitempty"`       // BT-31 / BT-48
	TaxRegistrationID   string   `json:"taxRegistrationId,omitempty"`   // BT-32 (Seller only)
	AdditionalLegalInfo string   `json:"additionalLegalInfo,omitempty"` // BT-33 (Seller only)
	ElectronicAddress   *ID      `json:"electronicAddress,omitempty"`   // BT-34(+1) / BT-49(+1) (scheme = EAS)
	Address             Address  `json:"address"`                       // BG-5 / BG-8
	Contact             *Contact `json:"contact,omitempty"`             // BG-6 / BG-9
}

// Address is a postal address. BG-5 / BG-8 / BG-12 / BG-15.
type Address struct {
	Line1       string `json:"line1,omitempty"`       // BT-35 / 50 / 64 / 75
	Line2       string `json:"line2,omitempty"`       // BT-36 / 51 / 65 / 76
	Line3       string `json:"line3,omitempty"`       // BT-162 / 163 / 164 / 165
	City        string `json:"city,omitempty"`        // BT-37 / 52 / 66 / 77
	PostCode    string `json:"postCode,omitempty"`    // BT-38 / 53 / 67 / 78
	Subdivision string `json:"subdivision,omitempty"` // BT-39 / 54 / 68 / 79
	CountryCode Code   `json:"countryCode"`           // BT-40 / 55 / 69 / 80
}

// Contact is a contact point. BG-6 / BG-9.
type Contact struct {
	Point string `json:"point,omitempty"` // BT-41 / 56 Contact point
	Phone string `json:"phone,omitempty"` // BT-42 / 57 Contact telephone number
	Email string `json:"email,omitempty"` // BT-43 / 58 Contact email address
}

// Payee is the party to whom payment is owed when different from the Seller. BG-10.
type Payee struct {
	Name                string `json:"name"`                          // BT-59
	Identifier          *ID    `json:"identifier,omitempty"`          // BT-60(+1)
	LegalRegistrationID *ID    `json:"legalRegistrationId,omitempty"` // BT-61(+1)
}

// TaxRepresentative is the Seller's tax representative party. BG-11 (+ BG-12).
type TaxRepresentative struct {
	Name          string  `json:"name"`          // BT-62
	VATIdentifier string  `json:"vatIdentifier"` // BT-63
	Address       Address `json:"address"`       // BG-12
}

// Delivery holds delivery information. BG-13 (+ BG-14 period, BG-15 address).
type Delivery struct {
	PartyName          string   `json:"partyName,omitempty"`          // BT-70 Deliver to party name
	LocationIdentifier *ID      `json:"locationIdentifier,omitempty"` // BT-71(+1)
	ActualDeliveryDate Date     `json:"actualDeliveryDate,omitempty"` // BT-72
	InvoicingPeriod    *Period  `json:"invoicingPeriod,omitempty"`    // BG-14 (BT-73/74)
	Address            *Address `json:"address,omitempty"`            // BG-15 (BT-75..80)
}

// Period is a start/end date range. BG-14 (invoicing period) / BG-26 (line period).
type Period struct {
	Start Date `json:"start,omitempty"` // BT-73 / BT-134
	End   Date `json:"end,omitempty"`   // BT-74 / BT-135
}

// PaymentInstructions describes how the invoice is to be paid. BG-16.
type PaymentInstructions struct {
	MeansTypeCode         Code             `json:"meansTypeCode,omitempty"`         // BT-81
	MeansText             string           `json:"meansText,omitempty"`             // BT-82
	RemittanceInformation string           `json:"remittanceInformation,omitempty"` // BT-83
	CreditTransfers       []CreditTransfer `json:"creditTransfers,omitempty"`       // BG-17 (0..n)
	Card                  *PaymentCard     `json:"card,omitempty"`                  // BG-18
	DirectDebit           *DirectDebit     `json:"directDebit,omitempty"`           // BG-19
}

// CreditTransfer is a credit-transfer payment account. BG-17.
type CreditTransfer struct {
	AccountIdentifier string `json:"accountIdentifier"`           // BT-84 (IBAN or proprietary)
	AccountName       string `json:"accountName,omitempty"`       // BT-85
	ServiceProviderID string `json:"serviceProviderId,omitempty"` // BT-86 (BIC)
}

// PaymentCard is payment-card information. BG-18.
type PaymentCard struct {
	PrimaryAccountNumber string `json:"primaryAccountNumber"` // BT-87
	HolderName           string `json:"holderName,omitempty"` // BT-88
}

// DirectDebit is direct-debit information. BG-19.
type DirectDebit struct {
	MandateReference   string `json:"mandateReference,omitempty"`   // BT-89
	CreditorIdentifier string `json:"creditorIdentifier,omitempty"` // BT-90
	DebitedAccount     string `json:"debitedAccount,omitempty"`     // BT-91
}

// AllowanceCharge is a document-level allowance (BG-20) or charge (BG-21). The
// same shape is used for both; the slice it lives in determines the sense.
type AllowanceCharge struct {
	Amount          Amount     `json:"amount"`               // BT-92 / BT-99
	BaseAmount      Amount     `json:"baseAmount,omitempty"` // BT-93 / BT-100
	Percentage      Percentage `json:"percentage,omitempty"` // BT-94 / BT-101
	VATCategoryCode Code       `json:"vatCategoryCode"`      // BT-95 / BT-102
	VATRate         Percentage `json:"vatRate,omitempty"`    // BT-96 / BT-103
	Reason          string     `json:"reason,omitempty"`     // BT-97 / BT-104
	ReasonCode      Code       `json:"reasonCode,omitempty"` // BT-98 / BT-105
}

// DocumentTotals are the document total amounts. BG-22.
type DocumentTotals struct {
	LineNetTotal               Amount `json:"lineNetTotal"`                         // BT-106 Sum of Invoice line net amount
	AllowanceTotal             Amount `json:"allowanceTotal,omitempty"`             // BT-107 Sum of allowances on document level
	ChargeTotal                Amount `json:"chargeTotal,omitempty"`                // BT-108 Sum of charges on document level
	TaxBasisTotal              Amount `json:"taxBasisTotal"`                        // BT-109 Invoice total amount without VAT
	TaxTotal                   Amount `json:"taxTotal,omitempty"`                   // BT-110 Invoice total VAT amount
	TaxTotalAccountingCurrency Amount `json:"taxTotalAccountingCurrency,omitempty"` // BT-111 Invoice total VAT amount in accounting currency
	GrandTotal                 Amount `json:"grandTotal"`                           // BT-112 Invoice total amount with VAT
	PaidAmount                 Amount `json:"paidAmount,omitempty"`                 // BT-113 Paid amount
	RoundingAmount             Amount `json:"roundingAmount,omitempty"`             // BT-114 Rounding amount
	DuePayableAmount           Amount `json:"duePayableAmount"`                     // BT-115 Amount due for payment
}

// VATBreakdown is one VAT category subtotal. BG-23 (1..n).
type VATBreakdown struct {
	TaxableAmount       Amount     `json:"taxableAmount"`                 // BT-116 VAT category taxable amount
	TaxAmount           Amount     `json:"taxAmount"`                     // BT-117 VAT category tax amount
	CategoryCode        Code       `json:"categoryCode"`                  // BT-118 VAT category code
	Rate                Percentage `json:"rate,omitempty"`                // BT-119 VAT category rate
	ExemptionReasonText string     `json:"exemptionReasonText,omitempty"` // BT-120 VAT exemption reason text
	ExemptionReasonCode Code       `json:"exemptionReasonCode,omitempty"` // BT-121 VAT exemption reason code
}

// SupportingDocument is an additional supporting document. BG-24.
type SupportingDocument struct {
	Reference        string        `json:"reference"`                  // BT-122 Supporting document reference
	Description      string        `json:"description,omitempty"`      // BT-123 Supporting document description
	ExternalLocation string        `json:"externalLocation,omitempty"` // BT-124 External document location
	Attachment       *BinaryObject `json:"attachment,omitempty"`       // BT-125 (+ BT-125-1 mime, BT-125-2 filename)
}

// Line is a single invoice line. BG-25 (1..n).
type Line struct {
	ID                       string                `json:"id"`                                 // BT-126 Invoice line identifier
	Note                     string                `json:"note,omitempty"`                     // BT-127 Invoice line note
	ObjectIdentifier         *ID                   `json:"objectIdentifier,omitempty"`         // BT-128(+1)
	Quantity                 Quantity              `json:"quantity"`                           // BT-129 (+ BT-130 unit)
	NetAmount                Amount                `json:"netAmount"`                          // BT-131 Invoice line net amount
	OrderLineReference       string                `json:"orderLineReference,omitempty"`       // BT-132
	BuyerAccountingReference string                `json:"buyerAccountingReference,omitempty"` // BT-133
	Period                   *Period               `json:"period,omitempty"`                   // BG-26 (BT-134/135)
	Allowances               []LineAllowanceCharge `json:"allowances,omitempty"`               // BG-27 (0..n)
	Charges                  []LineAllowanceCharge `json:"charges,omitempty"`                  // BG-28 (0..n)
	Price                    Price                 `json:"price"`                              // BG-29
	VAT                      LineVAT               `json:"vat"`                                // BG-30
	Item                     Item                  `json:"item"`                               // BG-31
	SubLines                 []Line                `json:"subLines,omitempty"`                 // BG-DEX-01 sub-invoice lines (XRechnung Extension, UBL only)
}

// LineAllowanceCharge is a line-level allowance (BG-27) or charge (BG-28). Unlike
// document-level allowances/charges these carry no VAT category of their own.
type LineAllowanceCharge struct {
	Amount     Amount     `json:"amount"`               // BT-136 / BT-141
	BaseAmount Amount     `json:"baseAmount,omitempty"` // BT-137 / BT-142
	Percentage Percentage `json:"percentage,omitempty"` // BT-138 / BT-143
	Reason     string     `json:"reason,omitempty"`     // BT-139 / BT-144
	ReasonCode Code       `json:"reasonCode,omitempty"` // BT-140 / BT-145
}

// Price is the line price details. BG-29.
type Price struct {
	NetPrice     Amount    `json:"netPrice"`               // BT-146 Item net price
	Discount     Amount    `json:"discount,omitempty"`     // BT-147 Item price discount
	GrossPrice   Amount    `json:"grossPrice,omitempty"`   // BT-148 Item gross price
	BaseQuantity *Quantity `json:"baseQuantity,omitempty"` // BT-149 (+ BT-150 unit)
}

// LineVAT is the line VAT information. BG-30. The exemption-reason fields are not
// EN16931 line terms but some invoices carry a reason on the line VAT; they are
// preserved here for round-trip fidelity.
type LineVAT struct {
	CategoryCode        Code       `json:"categoryCode"`                  // BT-151 Invoiced item VAT category code
	Rate                Percentage `json:"rate,omitempty"`                // BT-152 Invoiced item VAT rate
	ExemptionReasonText string     `json:"exemptionReasonText,omitempty"` // line-level VAT exemption reason
	ExemptionReasonCode Code       `json:"exemptionReasonCode,omitempty"` // line-level VAT exemption reason code
}

// Item is the line item (product/service) information. BG-31.
type Item struct {
	Name               string               `json:"name"`                         // BT-153 Item name
	Description        string               `json:"description,omitempty"`        // BT-154 Item description
	SellerIdentifier   string               `json:"sellerIdentifier,omitempty"`   // BT-155 Item Seller's identifier
	BuyerIdentifier    string               `json:"buyerIdentifier,omitempty"`    // BT-156 Item Buyer's identifier
	StandardIdentifier *ID                  `json:"standardIdentifier,omitempty"` // BT-157(+1)
	Classifications    []ItemClassification `json:"classifications,omitempty"`    // BT-158(+1 scheme, +2 version) (0..n)
	CountryOfOrigin    Code                 `json:"countryOfOrigin,omitempty"`    // BT-159 Item country of origin
	Attributes         []ItemAttribute      `json:"attributes,omitempty"`         // BG-32 (0..n)
}

// ItemClassification is a classification identifier for the item. BT-158.
type ItemClassification struct {
	Code        string `json:"code"`                  // BT-158 Item classification identifier
	ListID      Code   `json:"listId,omitempty"`      // BT-158-1 Item classification scheme identifier
	ListVersion string `json:"listVersion,omitempty"` // BT-158-2 Item classification scheme version identifier
}

// ItemAttribute is one item attribute name/value pair. BG-32 (BT-160/161).
type ItemAttribute struct {
	Name  string `json:"name"`  // BT-160 Item attribute name
	Value string `json:"value"` // BT-161 Item attribute value
}
