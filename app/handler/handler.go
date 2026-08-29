package handler

import (
	category "MoneyHook/MoneyHook-API/cagegory"
	fixed "MoneyHook/MoneyHook-API/fixed"
	job "MoneyHook/MoneyHook-API/job"
	payment_resource "MoneyHook/MoneyHook-API/payment_resource"
	"MoneyHook/MoneyHook-API/router"
	sub_category "MoneyHook/MoneyHook-API/sub_cagegory"
	transaction "MoneyHook/MoneyHook-API/transaction"
	user "MoneyHook/MoneyHook-API/user"
)

type Handler struct {
	firebaseClient       router.IDTokenVerifier
	userStore            user.Store
	transactionStore     transaction.Store
	fixedStore           fixed.Store
	categoryStore        category.Store
	subCategoryStore     sub_category.Store
	paymentResourceStore payment_resource.Store
	jobsStore            job.Store
}

func NewHandler(
	fc router.IDTokenVerifier,
	us user.Store,
	ts transaction.Store,
	fs fixed.Store,
	cs category.Store,
	scs sub_category.Store,
	pr payment_resource.Store,
	js job.Store,
) *Handler {
	return &Handler{
		firebaseClient:       fc,
		categoryStore:        cs,
		subCategoryStore:     scs,
		transactionStore:     ts,
		fixedStore:           fs,
		userStore:            us,
		paymentResourceStore: pr,
		jobsStore:            js,
	}
}
