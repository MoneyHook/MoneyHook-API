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

type transactionResponse struct {
	Transaction transactionData `json:"transaction"`
}

type transactionData struct {
	TransactionDate   string `json:"transaction_date"`
	TransactionName   string `json:"transaction_name"`
	TransactionAmount int    `json:"transaction_amount"`
	CategoryId        int    `json:"category_id"`
	CategoryName      string `json:"category_name"`
	SubCategoryId     int    `json:"sub_category_id"`
	SubCategoryName   string `json:"sub_category_name"`
	FixedFlg          bool   `json:"fixed_flg"`
}

func getTransactionResponse(data *model.TransactionData) *transactionResponse {
	tr := &transactionResponse{}

	tr.Transaction = transactionData{
		TransactionDate:   data.TransactionDate.Format("2006-01-02"),
		TransactionName:   data.TransactionName,
		TransactionAmount: data.TransactionAmount,
		CategoryId:        data.CategoryId,
		CategoryName:      data.CategoryName,
		SubCategoryId:     data.SubCategoryId,
		SubCategoryName:   data.SubCategoryName,
		FixedFlg:          data.FixedFlg}

	return tr
}

type montylyFixedReponse struct {
	DisposableIncome int                `json:"disposable_income"`
	MontylyFixedList []montylyFixedData `json:"monthly_fixed_list"`
}

type montylyFixedData struct {
	CategoryName        string                    `json:"category_name"`
	TotalCategoryAmount int                       `json:"total_category_amount"`
	TransactionList     []montylyFixedTransaction `json:"transaction_list"`
}

type montylyFixedTransaction struct {
	TransactionName   string `json:"transaction_name"`
	TransactionAmount int    `json:"transaction_amount"`
}

func getMonthlyFixedResponse(data *[]model.MonthlyFixedData) *montylyFixedReponse {
	mfir := &montylyFixedReponse{}

	mfid_l := &[]montylyFixedData{}
	for _, category := range *data {
		if containsMonthlyFixedList(mfid_l, &category.CategoryName) {
			continue
		}

		mfid := &montylyFixedData{CategoryName: category.CategoryName, TotalCategoryAmount: category.TotalCategoryAmount}
		for _, transaction := range *data {
			if mfid.CategoryName == transaction.CategoryName {
				mfid.TransactionList = append(mfid.TransactionList,
					montylyFixedTransaction{TransactionName: transaction.TransactionName, TransactionAmount: transaction.TransactionAmount})
			}
		}

		*mfid_l = append(*mfid_l, *mfid)

	}
	mfir.MontylyFixedList = *mfid_l

	return mfir
}

func containsMonthlyFixedList(data_list *[]montylyFixedData, category_name *string) bool {
	for _, v := range *data_list {
		if v.CategoryName == *category_name {
			return true
		}
	}
	return false
}

type homeResponse struct {
	Balance          int            `json:"balance"`
	HomeCategoryList []homeCategory `json:"category_list"`
}

type homeCategory struct {
	CategoryName         string            `json:"category_name"`
	CategoryTotoalAmount int               `json:"category_total_amount"`
	HomeSubCategoryList  []homeSubCategory `json:"sub_category_list"`
}

type homeSubCategory struct {
	SubCategoryName        string `json:"sub_category_name"`
	SubCategoryTotalAmount int    `json:"sub_category_total_amount"`
}

func getHomeResponse(data *[]model.HomeCategory) *homeResponse {
	hr := &homeResponse{}

	hcl := &[]homeCategory{}
	for _, category := range *data {
		if containsHomeList(hcl, &category.CategoryName) {
			continue
		}

		hc := &homeCategory{CategoryName: category.CategoryName, CategoryTotoalAmount: category.CategoryTotalAmount}

		for _, sub_category := range *data {
			if hc.CategoryName == sub_category.CategoryName {
				hc.HomeSubCategoryList = append(hc.HomeSubCategoryList,
					homeSubCategory{SubCategoryName: sub_category.SubCategoryName, SubCategoryTotalAmount: sub_category.SubCategoryTotalAmount})
			}
		}

		*hcl = append(*hcl, *hc)
		hr.Balance += category.CategoryTotalAmount
	}

	hr.HomeCategoryList = *hcl

	return hr
}

func containsHomeList(data_list *[]homeCategory, category_name *string) bool {
	for _, v := range *data_list {
		if v.CategoryName == *category_name {
			return true
		}
	}
	return false
}

type monthlyVariableResponse struct {
	TotalVariable       int                       `json:"total_variable"`
	MonthlyVariableList []monthlyVariableCategory `json:"monthly_variable_list"`
}

type monthlyVariableCategory struct {
	CategoryName               string                       `json:"category_name"`
	CategoryTotoalAmount       int                          `json:"category_total_amount"`
	MonthlyVariableSubCategory []monthlyVariableSubCategory `json:"sub_category_list"`
}

type monthlyVariableSubCategory struct {
	SubCategoryId          int                          `json:"sub_category_id"`
	SubCategoryName        string                       `json:"sub_category_name"`
	SubCategoryTotalAmount int                          `json:"sub_category_total_amount"`
	TransactionList        []monthlyVariableTransaction `json:"transaction_list"`
}

type monthlyVariableTransaction struct {
	TransactionId     int    `json:"transaction_id"`
	TransactionName   string `json:"transaction_name"`
	TransactionAmount int    `json:"transaction_amount"`
}

func getMonthlyVariableResponse(data *[]model.MonthlyVariableData) *monthlyVariableResponse {
	mvr := &monthlyVariableResponse{}

	mvcl := &[]monthlyVariableCategory{}
	for _, category := range *data {
		if containsVariableCategory(mvcl, &category.CategoryName) {
			continue
		}

		mvc := &monthlyVariableCategory{CategoryName: category.CategoryName, CategoryTotoalAmount: category.CategoryTotalAmount}

		for _, sub_category := range *data {
			if mvc.CategoryName == sub_category.CategoryName {
				mvsc := &monthlyVariableSubCategory{SubCategoryId: sub_category.SubCategoryId,
					SubCategoryName:        sub_category.SubCategoryName,
					SubCategoryTotalAmount: sub_category.SubCategoryTotalAmount}

				if containsVariableSubCategory(&mvc.MonthlyVariableSubCategory, &sub_category.SubCategoryName) {
					continue
				}
				for _, transaction := range *data {
					if mvsc.SubCategoryId == transaction.SubCategoryId {
						mvt := &monthlyVariableTransaction{TransactionId: transaction.TransactionId,
							TransactionName:   transaction.TransactionName,
							TransactionAmount: transaction.TransactionAmount}

						mvsc.TransactionList = append(mvsc.TransactionList, *mvt)
					}
				}

				mvc.MonthlyVariableSubCategory = append(mvc.MonthlyVariableSubCategory, *mvsc)
			}
		}
		*mvcl = append(*mvcl, *mvc)
		mvr.TotalVariable += category.CategoryTotalAmount
	}

	mvr.MonthlyVariableList = *mvcl

	return mvr
}

func containsVariableCategory(data_list *[]monthlyVariableCategory, category_name *string) bool {
	for _, v := range *data_list {
		if v.CategoryName == *category_name {
			return true
		}
	}
	return false
}

func containsVariableSubCategory(data_list *[]monthlyVariableSubCategory, sub_category_name *string) bool {
	for _, v := range *data_list {
		if v.SubCategoryName == *sub_category_name {
			return true
		}
	}
	return false
}

type totalSpendingResponse struct {
	TotalSpending     int                     `json:"total_spending"`
	TotalSpendingList []totalSpendingCategory `json:"category_total_list"`
}

type totalSpendingCategory struct {
	CategoryName             string                     `json:"category_name"`
	CategoryTotoalAmount     int                        `json:"category_total_amount"`
	TotalSpendingSubCategory []totalSpendingSubCategory `json:"sub_category_list"`
}

type totalSpendingSubCategory struct {
	SubCategoryId          int                        `json:"sub_category_id"`
	SubCategoryName        string                     `json:"sub_category_name"`
	SubCategoryTotalAmount int                        `json:"sub_category_total_amount"`
	TransactionList        []totalSpendingTransaction `json:"transaction_list"`
}

type totalSpendingTransaction struct {
	TransactionId     int    `json:"transaction_id"`
	TransactionName   string `json:"transaction_name"`
	TransactionAmount int    `json:"transaction_amount"`
}

func getTotalSpendingResponse(data *[]model.TotalSpendingData) *totalSpendingResponse {
	tsr := &totalSpendingResponse{}

	mvcl := &[]totalSpendingCategory{}
	for _, category := range *data {
		if containsTotalSpendingCategory(mvcl, &category.CategoryName) {
			continue
		}

		mvc := &totalSpendingCategory{CategoryName: category.CategoryName, CategoryTotoalAmount: category.CategoryTotalAmount}

		for _, sub_category := range *data {
			if mvc.CategoryName == sub_category.CategoryName {
				mvsc := &totalSpendingSubCategory{SubCategoryId: sub_category.SubCategoryId,
					SubCategoryName:        sub_category.SubCategoryName,
					SubCategoryTotalAmount: sub_category.SubCategoryTotalAmount}

				if containsTotalSpendingSubCategory(&mvc.TotalSpendingSubCategory, &sub_category.SubCategoryName) {
					continue
				}
				for _, transaction := range *data {
					if mvsc.SubCategoryId == transaction.SubCategoryId {
						mvt := &totalSpendingTransaction{TransactionId: transaction.TransactionId,
							TransactionName:   transaction.TransactionName,
							TransactionAmount: transaction.TransactionAmount}

						mvsc.TransactionList = append(mvsc.TransactionList, *mvt)
					}
				}

				mvc.TotalSpendingSubCategory = append(mvc.TotalSpendingSubCategory, *mvsc)
			}
		}
		*mvcl = append(*mvcl, *mvc)
		tsr.TotalSpending += category.CategoryTotalAmount
	}

	tsr.TotalSpendingList = *mvcl

	return tsr
}

func containsTotalSpendingCategory(data_list *[]totalSpendingCategory, category_name *string) bool {
	for _, v := range *data_list {
		if v.CategoryName == *category_name {
			return true
		}
	}
	return false
}

func containsTotalSpendingSubCategory(data_list *[]totalSpendingSubCategory, sub_category_name *string) bool {
	for _, v := range *data_list {
		if v.SubCategoryName == *sub_category_name {
			return true
		}
	}
	return false
}

type fixedResponse struct {
	MonthlyTransactionList []fixedResponseData `json:"monthly_transaction_list"`
}

type fixedResponseData struct {
	MonthlyTransactionId     int    `json:"monthly_transaction_id"`
	MonthlyTransactionName   string `json:"monthly_transaction_name"`
	MonthlyTransactionAmount int    `json:"monthly_transaction_amount"`
	MonthlyTransactionSign   int    `json:"monthly_transaction_sign"`
	MonthlyTransactionDate   int    `json:"monthly_transaction_date"`
	CategoryId               int    `json:"category_id"`
	CategoryName             string `json:"category_name"`
	SubCategoryId            int    `json:"sub_category_id"`
	SubCategoryName          string `json:"sub_category_name"`
}

func GetFixedResponse(data_list *[]model.GetFixed) *fixedResponse {
	fr := &[]fixedResponseData{}

	for _, data := range *data_list {
		*fr = append(*fr,
			fixedResponseData{MonthlyTransactionId: data.MonthlyTransactionId,
				MonthlyTransactionName:   data.MonthlyTransactionName,
				MonthlyTransactionAmount: data.MonthlyTransactionAmount,
				MonthlyTransactionSign:   data.MonthlyTransactionSign,
				MonthlyTransactionDate:   data.MonthlyTransactionDate,
				CategoryId:               data.CategoryId,
				CategoryName:             data.CategoryName,
				SubCategoryId:            data.SubCategoryId,
				SubCategoryName:          data.SubCategoryName,
			})
	}

	return &fixedResponse{MonthlyTransactionList: *fr}
}

type deletedFixedResponse struct {
	MonthlyTransactionId     int    `json:"monthly_transaction_id"`
	MonthlyTransactionName   string `json:"monthly_transaction_name"`
	MonthlyTransactionAmount int    `json:"monthly_transaction_amount"`
	MonthlyTransactionDate   int    `json:"monthly_transaction_date"`
	CategoryName             string `json:"category_name"`
	SubCategoryName          string `json:"sub_category_name"`
}

func GetFixedDeletedResponse(data_list *[]model.GetDeletedFixed) *[]deletedFixedResponse {
	dfr := &[]deletedFixedResponse{}

	for _, data := range *data_list {
		*dfr = append(*dfr,
			deletedFixedResponse{MonthlyTransactionId: data.MonthlyTransactionId,
				MonthlyTransactionName:   data.MonthlyTransactionName,
				MonthlyTransactionAmount: data.MonthlyTransactionAmount,
				MonthlyTransactionDate:   data.MonthlyTransactionDate,
				CategoryName:             data.CategoryName,
				SubCategoryName:          data.SubCategoryName,
			})
	}

	return dfr
}
