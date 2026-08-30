package handler

import (
	categorydomain "MoneyHook/MoneyHook-API/category"
	fixeddomain "MoneyHook/MoneyHook-API/fixed"
	analyticshandler "MoneyHook/MoneyHook-API/handler/analytics"
	categoryhandler "MoneyHook/MoneyHook-API/handler/category"
	fixedhandler "MoneyHook/MoneyHook-API/handler/fixed"
	jobhandler "MoneyHook/MoneyHook-API/handler/job"
	paymenthandler "MoneyHook/MoneyHook-API/handler/payment"
	subcategoryhandler "MoneyHook/MoneyHook-API/handler/subcategory"
	transactionhandler "MoneyHook/MoneyHook-API/handler/transaction"
	jobdomain "MoneyHook/MoneyHook-API/job"
	paymentresource "MoneyHook/MoneyHook-API/paymentresource"
	"MoneyHook/MoneyHook-API/router"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"MoneyHook/MoneyHook-API/user"
)

type Dependencies struct {
	FirebaseClient       router.IDTokenVerifier
	UserStore            user.Store
	TransactionStore     transactiondomain.Store
	FixedStore           fixeddomain.Store
	CategoryStore        categorydomain.Store
	SubCategoryStore     subcategorydomain.Store
	PaymentResourceStore paymentresource.Store
	JobStore             jobdomain.Store
}

type Handler struct {
	firebaseClient router.IDTokenVerifier
	userStore      user.Store
	transaction    *transactionhandler.Handler
	analytics      *analyticshandler.Handler
	fixed          *fixedhandler.Handler
	category       *categoryhandler.Handler
	subcategory    *subcategoryhandler.Handler
	payment        *paymenthandler.Handler
	job            *jobhandler.Handler
}

func New(dependencies Dependencies) *Handler {
	return &Handler{
		firebaseClient: dependencies.FirebaseClient,
		userStore:      dependencies.UserStore,
		transaction: transactionhandler.New(
			dependencies.TransactionStore,
			dependencies.SubCategoryStore,
			dependencies.PaymentResourceStore,
		),
		analytics:   analyticshandler.New(dependencies.TransactionStore),
		fixed:       fixedhandler.New(dependencies.FixedStore, dependencies.SubCategoryStore),
		category:    categoryhandler.New(dependencies.CategoryStore),
		subcategory: subcategoryhandler.New(dependencies.SubCategoryStore),
		payment:     paymenthandler.New(dependencies.PaymentResourceStore),
		job:         jobhandler.New(dependencies.JobStore),
	}
}
