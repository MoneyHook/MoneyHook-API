package model

type PaymentResource struct {
	PaymentId   int
	PaymentName string
}

type AddPaymentResource struct {
	UserNo      int
	PaymentName string
}

type DeletePaymentResource struct {
	UserNo    int
	PaymentId int
}
