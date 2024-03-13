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
	CategoryId              int                                 `json:"category_id"`
	CategoryName            string                              `json:"category_name"`
	SubCategoryListResponse []subCategoryListWithEnableResponse `json:"sub_category_list"`
}

type subCategoryListWithEnableResponse struct {
	SubCategoryId   int    `json:"sub_category_id"`
	SubCategoryName string `json:"sub_category_name"`
	Enable          bool   `json:"enable"`
}

func getCategoryWithSubCategoryListResponse(data *[]model.CategoryWithSubCategory) *categoryWithSubCategoryListResponse {
	cl := &categoryWithSubCategoryListResponse{}

	for _, category := range *data {
		scl := []subCategoryListWithEnableResponse{}
		for _, sub_category := range category.SubCategoryList {
			scl = append(scl, subCategoryListWithEnableResponse{SubCategoryId: sub_category.SubCategoryId, SubCategoryName: sub_category.SubCategoryName, Enable: sub_category.Enable})
		}

		cr := &categoryWithSubCategory{CategoryId: category.CategoryId, CategoryName: category.CategoryName, SubCategoryListResponse: scl}
		cl.CategoryList = append(cl.CategoryList, *cr)
	}

	return cl
}

type subCategoryListResponse struct {
	SubCategoryList []subCategoryResponse `json:"sub_category_list"`
}

type subCategoryResponse struct {
	SubCategoryId   int    `json:"sub_category_id"`
	SubCategoryName string `json:"sub_category_name"`
}

func getSubCategoryListResponse(data *[]model.SubCategory) *subCategoryListResponse {
	scl := &subCategoryListResponse{}

	for _, sub_category := range *data {
		scr := &subCategoryResponse{SubCategoryId: sub_category.SubCategoryId, SubCategoryName: sub_category.SubCategoryName}
		scl.SubCategoryList = append(scl.SubCategoryList, *scr)
	}

	return scl
}

type timelineListResponse struct {
	TimelineList []timelineResponse `json:"transaction_list"`
}

type timelineResponse struct {
	TransactionId     int    `json:"transaction_id"`
	TransactionName   string `json:"transaction_name"`
	TransactionAmount int    `json:"transaction_amount"`
	TransactionSign   int    `json:"transaction_sign"`
	TransactionDate   string `json:"transaction_date"`
	CategoryId        int    `json:"category_id"`
	CategoryName      string `json:"category_name"`
	SubCategoryId     int    `json:"sub_category_id"`
	SubCategoryName   string `json:"sub_category_name"`
	FixedFlg          bool   `json:"fixed_flg"`
}

func getTimelineListResponse(data *[]model.Timeline) *timelineListResponse {
	tll := &timelineListResponse{}

	for _, t := range *data {
		tl := &timelineResponse{TransactionId: t.TransactionId,
			TransactionName:   t.TransactionName,
			TransactionAmount: t.TransactionAmount,
			TransactionSign:   t.TransactionSign,
			TransactionDate:   t.TransactionDate.Format("2006-01-02"),
			CategoryId:        t.CategoryId,
			CategoryName:      t.CategoryName,
			SubCategoryId:     t.SubCategoryId,
			SubCategoryName:   t.SubCategoryName,
			FixedFlg:          t.FixedFlg}
		tll.TimelineList = append(tll.TimelineList, *tl)
	}

	return tll
}

type monthlySpendingDataResponse struct {
	MonthlyTotalAmountList []monthlyTotalAmount `json:"monthly_total_amount_list"`
}

type monthlyTotalAmount struct {
	TotalAmount int    `json:"total_amount"`
	Month       string `json:"month"`
}

func getmonthlySpendingDataResponse(data *[]model.MonthlySpendingData) *monthlySpendingDataResponse {
	msdr := &monthlySpendingDataResponse{}

	for _, m := range *data {
		ml := &monthlyTotalAmount{TotalAmount: m.TotalAmount, Month: m.Month}
		msdr.MonthlyTotalAmountList = append(msdr.MonthlyTotalAmountList, *ml)
	}
	return msdr
}
