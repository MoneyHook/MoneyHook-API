package handler

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Register(v1 *echo.Group) {
	category := v1.Group("/category")
	category.GET("/getCategoryList", h.GetCategoryList)
	category.GET("/getCategoryWithSubCategoryList", h.GetCategoryWithSubCategoryList)
}
