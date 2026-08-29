package handler

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(v1 *echo.Group) {
	reactV1 := v1.Group("/v1")
	v1Transactions := reactV1.Group("/transactions")
	v1Transactions.GET("/:transactionId", h.getV1Transaction)
	v1Transactions.POST("", h.createV1Transaction)
	v1Transactions.PATCH("/:transactionId", h.updateV1Transaction)
	v1Transactions.DELETE("/:transactionId", h.deleteV1Transaction)
	v1Analytics := reactV1.Group("/analytics")
	v1Analytics.GET("/overview", h.getV1AnalyticsOverview)
	v1Analytics.GET("/categories", h.getV1AnalyticsCategories)
	v1Analytics.GET("/fixed", h.getV1AnalyticsFixed)
	v1Analytics.GET("/payments", h.getV1AnalyticsPayments)

	user := v1.Group("/user")
	user.POST("/googleSignIn", h.googleSignIn)

	transaction := v1.Group("/transaction")
	transaction.GET("/getTimelineData", h.getTimelineData)
	transaction.GET("/getMonthlySpendingData", h.getMonthlySpendingData)
	transaction.GET("/getTransaction/:transactionId", h.getTransaction)
	transaction.GET("/getMonthlyFixedIncome", h.getMonthlyFixedIncome)
	transaction.GET("/getMonthlyFixedSpending", h.getMonthlyFixedSpending)
	transaction.GET("/getHome", h.getHome)
	transaction.GET("/getMonthlyVariableData", h.getMonthlyVariableData)
	transaction.GET("/getTotalSpending", h.getTotalSpendingData)
	transaction.GET("/groupByPayment", h.groupByPayment)
	transaction.GET("/getMonthlyWithdrawalAmount", h.getMonthlyWithdrawalAmount)
	transaction.GET("/getFrequentTransactionName", h.getFrequentTransactionName)
	transaction.POST("/addTransaction", h.addTransaction)
	transaction.POST("/addTransactionList", h.addTransactionList)
	transaction.PATCH("/editTransaction", h.editTransaction)
	transaction.DELETE("/deleteTransaction/:transactionId", h.deleteTransaction)

	fixed := v1.Group("/fixed")
	fixed.GET("/getFixed", h.getFixed)
	fixed.GET("/getDeletedFixed", h.getDeletedFixed)
	fixed.POST("/addFixed", h.addFixed)
	fixed.PATCH("/editFixed", h.editFixed)
	fixed.DELETE("/deleteFixed/:monthly_transaction_id", h.deleteFixed)

	category := v1.Group("/category")
	category.GET("/getCategoryList", h.GetCategoryList)
	category.GET("/getCategoryWithSubCategoryList", h.GetCategoryWithSubCategoryList)

	sub_category := v1.Group("/subCategory")
	sub_category.GET("/getSubCategoryList/:categoryId", h.GetSubCategoryList)
	sub_category.POST("/editSubCategory", h.EditSubCategory)

	payment_resource := v1.Group("/payment")
	payment_resource.GET("/getPayment", h.GetPaymentResourceList)
	payment_resource.POST("/addPayment", h.AddPaymentResource)
	payment_resource.PATCH("/editPayment", h.EditPaymentResource)
	payment_resource.DELETE("/deletePayment/:paymentId", h.DeletePaymentResource)
	payment_resource.GET("/getPaymentType", h.GetPaymentTypeList)

	job := v1.Group("/job")
	job.POST("/daily", h.ProcessDailyJob)
}
