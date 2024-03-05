package handler

import (
	"MoneyHook/MoneyHook-API/model"
)

type categoryResponse struct {
	Category_id   int    `json:"category_id"`
	Category_name string `json:"category_name"`
}

type categoryListResponse struct {
	CategoryList []categoryResponse `json:"category_list"`
}

func getCategoryListResponse(data *[]model.Category) *categoryListResponse {
	cl := &categoryListResponse{}

	for _, category := range *data {
		cr := &categoryResponse{Category_id: category.CategoryId, Category_name: category.CategoryName}
		cl.CategoryList = append(cl.CategoryList, *cr)
	}

	return cl
}

type categoryWithSubCategoryListResponse struct {
	CategoryList []categoryWithSubCategory `json:"category_list"`
}

type categoryWithSubCategory struct {
	CategoryId              int                       `json:"category_id"`
	CategoryName            string                    `json:"category_name"`
	SubCategoryListResponse []subCategoryListResponse `json:"sub_category_list"`
}

type subCategoryListResponse struct {
	SubCategoryId   int    `json:"sub_category_id"`
	SubCategoryName string `json:"sub_category_name"`
	Enable          bool   `json:"enable"`
}

func getCategoryWithSubCategoryListResponse(data *[]model.CategoryWithSubCategory) *categoryWithSubCategoryListResponse {
	cl := &categoryWithSubCategoryListResponse{}

	for _, category := range *data {
		scl := []subCategoryListResponse{}
		for _, sub_category := range category.SubCategoryList {
			scl = append(scl, subCategoryListResponse{SubCategoryId: sub_category.SubCategoryId, SubCategoryName: sub_category.SubCategoryName, Enable: sub_category.Enable})
		}

		cr := &categoryWithSubCategory{CategoryId: category.CategoryId, CategoryName: category.CategoryName, SubCategoryListResponse: scl}
		cl.CategoryList = append(cl.CategoryList, *cr)
	}

	return cl
}
