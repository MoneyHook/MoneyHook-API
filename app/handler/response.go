package handler

import (
	"example.com/m/model"
)

type categoryResponse struct {
	Category_id   int    `json:"categoryId"`
	Category_name string `json:"categoryName"`
}

type categoryListResponse struct {
	CategoryList []categoryResponse `json:"categoryList"`
}

func getCategoryListResponse(data *model.CategoryList) *categoryListResponse {
	cl := &categoryListResponse{}

	for _, category := range data.CategoryList {
		cr := &categoryResponse{Category_id: category.CategoryId, Category_name: category.CategoryName}
		cl.CategoryList = append(cl.CategoryList, *cr)
	}

	return cl
}

type categoryWithSubCategoryListResponse struct {
	CategoryList []categoryWithSubCategory `json:"categoryList"`
}

type categoryWithSubCategory struct {
	Category_id             int                       `json:"categoryId"`
	Category_name           string                    `json:"categoryName"`
	SubCategoryListResponse []subCategoryListResponse `json:"subCategoryList"`
}

type subCategoryListResponse struct {
	Sub_Category_id   int    `json:"sub_categoryId"`
	Sub_Category_name string `json:"sub_categoryName"`
	Enable            bool   `json:"enable"`
}

func getCategoryWithSubCategoryListResponse(data *model.CategoryWithSubCategoryList) *categoryWithSubCategoryListResponse {
	cl := &categoryWithSubCategoryListResponse{}

	for _, category := range data.CategoryWithSubCategoryList {
		scl := []subCategoryListResponse{}
		for _, sub_category := range category.SubCategoryList {
			scl = append(scl, subCategoryListResponse{Sub_Category_id: sub_category.SubCategoryId, Sub_Category_name: sub_category.SubCategoryName})
		}

		cr := &categoryWithSubCategory{Category_id: category.CategoryId, Category_name: category.CategoryName, SubCategoryListResponse: scl}
		cl.CategoryList = append(cl.CategoryList, *cr)
	}

	return cl
}
