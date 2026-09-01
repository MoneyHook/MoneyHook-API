package subcategory

import "MoneyHook/MoneyHook-API/model"

type subCategoryListResponse struct {
	SubCategoryList []subCategoryResponse `json:"sub_category_list"`
}

type subCategoryResponse struct {
	SubCategoryId   string `json:"sub_category_id"`
	SubCategoryName string `json:"sub_category_name"`
}

func getSubCategoryListResponse(data *[]model.SubCategory) *subCategoryListResponse {
	response := &subCategoryListResponse{}
	for _, subCategory := range *data {
		response.SubCategoryList = append(response.SubCategoryList, subCategoryResponse{
			SubCategoryId:   subCategory.SubCategoryId,
			SubCategoryName: subCategory.SubCategoryName,
		})
	}
	return response
}
