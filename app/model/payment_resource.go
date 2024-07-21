package model

type PaymentResource struct {
	PaymentId     int
	PaymentName   string
	PaymentTypeId int
	PaymentDate   int
	ClosingDate   int
}

type AddPaymentResource struct {
	UserNo        int
	PaymentName   string
	PaymentTypeId int
	PaymentDate   *int
	ClosingDate   *int
}

type EditPaymentResource struct {
	UserNo        int
	PaymentId     int
	PaymentName   string
	PaymentTypeId int
	PaymentDate   *int
	ClosingDate   *int
}

type DeletePaymentResource struct {
	UserNo    int
	PaymentId int
}

type PaymentType struct {
	PaymentTypeId     int
	PaymentTypeName   string
	IsPaymentDueLater bool
}
