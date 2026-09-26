package handler

import (
	budgetdomain "MoneyHook/MoneyHook-API/budget"
	categorydomain "MoneyHook/MoneyHook-API/category"
	fixeddomain "MoneyHook/MoneyHook-API/fixed"
	analyticshandler "MoneyHook/MoneyHook-API/handler/analytics"
	budgethandler "MoneyHook/MoneyHook-API/handler/budget"
	categoryhandler "MoneyHook/MoneyHook-API/handler/category"
	fixedhandler "MoneyHook/MoneyHook-API/handler/fixed"
	jobhandler "MoneyHook/MoneyHook-API/handler/job"
	paymenthandler "MoneyHook/MoneyHook-API/handler/payment"
	settingshandler "MoneyHook/MoneyHook-API/handler/settings"
	subcategoryhandler "MoneyHook/MoneyHook-API/handler/subcategory"
	transactionhandler "MoneyHook/MoneyHook-API/handler/transaction"
	jobdomain "MoneyHook/MoneyHook-API/job"
	paymentresource "MoneyHook/MoneyHook-API/paymentresource"
	"MoneyHook/MoneyHook-API/router"
	settingsdomain "MoneyHook/MoneyHook-API/settings"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"MoneyHook/MoneyHook-API/user"
)

type Dependencies struct {
	FirebaseClient       router.IDTokenVerifier
	SchedulerToken       router.SchedulerIDTokenValidator
	SchedulerAuth        router.SchedulerAuthConfig
	UserStore            user.Store
	BudgetStore          budgetdomain.Store
	SettingsStore        settingsdomain.Store
	TransactionStore     transactiondomain.Store
	FixedStore           fixeddomain.Store
	CategoryStore        categorydomain.Store
	SubCategoryStore     subcategorydomain.Store
	PaymentResourceStore paymentresource.Store
	JobStore             jobdomain.Store
}

type Handler struct {
	firebaseClient router.IDTokenVerifier
	schedulerToken router.SchedulerIDTokenValidator
	schedulerAuth  router.SchedulerAuthConfig
	userStore      user.Store
	budget         *budgethandler.Handler
	settings       *settingshandler.Handler
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
		schedulerToken: dependencies.SchedulerToken,
		schedulerAuth:  dependencies.SchedulerAuth,
		userStore:      dependencies.UserStore,
		budget:         budgethandler.New(dependencies.BudgetStore),
		settings:       settingshandler.New(dependencies.SettingsStore),
		transaction: transactionhandler.New(
			dependencies.TransactionStore,
			dependencies.PaymentResourceStore,
		),
		analytics:   analyticshandler.New(dependencies.TransactionStore),
		fixed:       fixedhandler.New(dependencies.FixedStore),
		category:    categoryhandler.New(dependencies.CategoryStore),
		subcategory: subcategoryhandler.New(dependencies.SubCategoryStore),
		payment:     paymenthandler.New(dependencies.PaymentResourceStore),
		job:         jobhandler.New(dependencies.JobStore, dependencies.SchedulerAuth.JobName),
	}
}
