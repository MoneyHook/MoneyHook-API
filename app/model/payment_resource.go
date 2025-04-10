package model

type PaymentResource struct {
	PaymentId     string
	PaymentName   string
	PaymentTypeId string
	PaymentDate   int
	ClosingDate   int
}

type AddPaymentResource struct {
	UserNo        string
	PaymentName   string
	PaymentTypeId string
	PaymentDate   *int
	ClosingDate   *int
}

type EditPaymentResource struct {
	UserNo        string
	PaymentId     string
	PaymentName   string
	PaymentTypeId string
	PaymentDate   *int
	ClosingDate   *int
}

type DeletePaymentResource struct {
	UserNo    string
	PaymentId string
}

type PaymentType struct {
	PaymentTypeId     string
	PaymentTypeName   string
	IsPaymentDueLater bool
}
