package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetCategoryList(c echo.Context) error {
	result := h.categoryStore.GetCategoryList()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	result_list := getCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) GetCategoryWithSubCategoryList(c echo.Context) error {
	result := h.categoryStore.GetCategoryWithSubCategoryList()
	// result := h.categoryStore.GetCategoryWithSubCategoryList()

	// fmt.Println(*result)
	// for _, c := range result.CategoryWithSubCategoryList {
	// 	fmt.Println(c.SubCategoryList)
	// }
	result_list := getCategoryWithSubCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}
