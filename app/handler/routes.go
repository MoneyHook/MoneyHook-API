package handler

import (
	"MoneyHook/MoneyHook-API/router"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(api *echo.Group) {
	authMiddleware := router.FirebaseAuthMiddleware(h.firebaseClient, h.userStore)
	protected := api.Group("", authMiddleware)

	reactV1 := protected.Group("/v1")
	budget := reactV1.Group("/budget")
	budget.GET("", h.budget.GetV1Budget)
	budget.PUT("", h.budget.PutV1Budget)
	settings := reactV1.Group("/settings")
	settings.GET("", h.settings.GetV1Settings)
	settings.PATCH("", h.settings.PatchV1Settings)
	v1Transactions := reactV1.Group("/transactions")
	v1Transactions.GET("/:transactionId", h.transaction.GetV1Transaction)
	v1Transactions.POST("", h.transaction.CreateV1Transaction)
	v1Transactions.PATCH("/:transactionId", h.transaction.UpdateV1Transaction)
	v1Transactions.DELETE("/:transactionId", h.transaction.DeleteV1Transaction)
	v1Analytics := reactV1.Group("/analytics")
	v1Analytics.GET("/overview", h.analytics.GetV1AnalyticsOverview)
	v1Analytics.GET("/categories", h.analytics.GetV1AnalyticsCategories)
	v1Analytics.GET("/fixed", h.analytics.GetV1AnalyticsFixed)
	v1Analytics.GET("/payments", h.analytics.GetV1AnalyticsPayments)

	transaction := protected.Group("/transaction")
	transaction.GET("/getTimelineData", h.transaction.GetTimelineData)
	transaction.GET("/getMonthlySpendingData", h.transaction.GetMonthlySpendingData)
	transaction.GET("/getTransaction/:transactionId", h.transaction.GetTransaction)
	transaction.GET("/getMonthlyFixedIncome", h.transaction.GetMonthlyFixedIncome)
	transaction.GET("/getMonthlyFixedSpending", h.transaction.GetMonthlyFixedSpending)
	transaction.GET("/getHome", h.transaction.GetHome)
	transaction.GET("/getMonthlyVariableData", h.transaction.GetMonthlyVariableData)
	transaction.GET("/getTotalSpending", h.transaction.GetTotalSpendingData)
	transaction.GET("/groupByPayment", h.transaction.GroupByPayment)
	transaction.GET("/getMonthlyWithdrawalAmount", h.transaction.GetMonthlyWithdrawalAmount)
	transaction.GET("/getFrequentTransactionName", h.transaction.GetFrequentTransactionName)
	transaction.POST("/addTransaction", h.transaction.AddTransaction)
	transaction.POST("/addTransactionList", h.transaction.AddTransactionList)
	transaction.PATCH("/editTransaction", h.transaction.EditTransaction)
	transaction.DELETE("/deleteTransaction/:transactionId", h.transaction.DeleteTransaction)

	fixed := protected.Group("/fixed")
	fixed.GET("/getFixed", h.fixed.GetFixed)
	fixed.GET("/getDeletedFixed", h.fixed.GetDeletedFixed)
	fixed.POST("/addFixed", h.fixed.AddFixed)
	fixed.PATCH("/editFixed", h.fixed.EditFixed)
	fixed.DELETE("/deleteFixed/:monthly_transaction_id", h.fixed.DeleteFixed)

	category := protected.Group("/category")
	category.GET("/getCategoryList", h.category.GetCategoryList)
	category.GET("/getCategoryWithSubCategoryList", h.category.GetCategoryWithSubCategoryList)

	subCategory := protected.Group("/subCategory")
	subCategory.GET("/getSubCategoryList/:categoryId", h.subcategory.GetSubCategoryList)
	subCategory.POST("/editSubCategory", h.subcategory.EditSubCategory)

	payment := protected.Group("/payment")
	payment.GET("/getPayment", h.payment.GetPaymentResourceList)
	payment.POST("/addPayment", h.payment.AddPaymentResource)
	payment.PATCH("/editPayment", h.payment.EditPaymentResource)
	payment.PUT("/reorder", h.payment.ReorderPaymentResources)
	payment.DELETE("/deletePayment/:paymentId", h.payment.DeletePaymentResource)
	payment.GET("/getPaymentType", h.payment.GetPaymentTypeList)

	job := protected.Group("/job")
	job.POST("/daily", h.job.ProcessDailyJob)
}
