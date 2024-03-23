package handler

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(v1 *echo.Group) {
	transaction := v1.Group("/transaction")
	transaction.GET("/getTimelineData", h.getTimelineData)
	transaction.GET("/getMonthlySpendingData", h.getMonthlySpendingData)
	transaction.GET("/getTransaction/:transactionId", h.getTransaction)
	transaction.GET("/getMonthlyFixedIncome", h.getMonthlyFixedIncome)
	transaction.GET("/getMonthlyFixedSpending", h.getMonthlyFixedSpending)
	transaction.GET("/getHome", h.getHome)

	category := v1.Group("/category")
	category.GET("/getCategoryList", h.GetCategoryList)
	category.GET("/getCategoryWithSubCategoryList", h.GetCategoryWithSubCategoryList)

	sub_category := v1.Group("/subCategory")
	sub_category.GET("/getSubCategoryList/:categoryId", h.GetSubCategoryList)
}
