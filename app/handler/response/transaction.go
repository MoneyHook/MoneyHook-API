package response

import (
	"MoneyHook/MoneyHook-API/model"
	"math"
)

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
	PaymentId         *int   `json:"payment_id"`
	PaymentName       string `json:"payment_name"`
}

func GetTimelineListResponse(data *[]model.Timeline) *timelineListResponse {
	tll := &timelineListResponse{TimelineList: []timelineResponse{}}

	for _, t := range *data {
		var paymentId *int
		if t.PaymentId != 0 {
			paymentId = &t.PaymentId
		}

		tl := &timelineResponse{TransactionId: t.TransactionId,
			TransactionName:   t.TransactionName,
			TransactionAmount: t.TransactionAmount,
			TransactionSign:   t.TransactionSign,
			TransactionDate:   t.TransactionDate.Format("2006-01-02"),
			CategoryId:        t.CategoryId,
			CategoryName:      t.CategoryName,
			SubCategoryId:     t.SubCategoryId,
			SubCategoryName:   t.SubCategoryName,
			FixedFlg:          t.FixedFlg,
			PaymentId:         paymentId,
			PaymentName:       t.PaymentName}
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

func GetmonthlySpendingDataResponse(data *[]model.MonthlySpendingData) *monthlySpendingDataResponse {
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

func GetTransactionResponse(data *model.TransactionData) *transactionResponse {
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
	CategoryId          int                       `json:"category_id"`
	CategoryName        string                    `json:"category_name"`
	TotalCategoryAmount int                       `json:"total_category_amount"`
	TransactionList     []montylyFixedTransaction `json:"transaction_list"`
}

type montylyFixedTransaction struct {
	TransactionId     int    `json:"transaction_id"`
	TransactionName   string `json:"transaction_name"`
	TransactionAmount int    `json:"transaction_amount"`
	TransactionDate   string `json:"transaction_date"`
	SubCategoryId     int    `json:"sub_category_id"`
	SubCategoryName   string `json:"sub_category_name"`
	FixedFlg          bool   `json:"fixed_flg"`
	PaymentId         int    `json:"payment_id"`
	PaymentName       string `json:"payment_name"`
}

func GetMonthlyFixedResponse(data *[]model.MonthlyFixedData) *montylyFixedReponse {
	mfir := &montylyFixedReponse{}

	mfid_l := &[]montylyFixedData{}
	for _, category := range *data {
		if containsMonthlyFixedList(mfid_l, &category.CategoryName) {
			continue
		}

		mfid := &montylyFixedData{CategoryId: category.CategoryId, CategoryName: category.CategoryName, TotalCategoryAmount: category.TotalCategoryAmount}
		for _, transaction := range *data {
			if mfid.CategoryName == transaction.CategoryName {
				mfid.TransactionList = append(mfid.TransactionList,
					montylyFixedTransaction{
						TransactionId:     transaction.TransactionId,
						TransactionName:   transaction.TransactionName,
						TransactionAmount: transaction.TransactionAmount,
						TransactionDate:   transaction.TransactionDate.Format("2006-01-02"),
						SubCategoryId:     transaction.SubCategoryId,
						SubCategoryName:   transaction.SubCategoryName,
						FixedFlg:          transaction.FixedFlg,
						PaymentId:         transaction.PaymentId,
						PaymentName:       transaction.PaymentName})
			}
		}

		*mfid_l = append(*mfid_l, *mfid)
		mfir.DisposableIncome += category.TotalCategoryAmount

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

func GetHomeResponse(data *[]model.HomeCategory) *homeResponse {
	hr := &homeResponse{HomeCategoryList: []homeCategory{}}

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

	other_hc := &homeCategory{CategoryName: "その他"}
	// 合計８カテゴリに統合
	for i, hc := range *hcl {
		if i > 6 {
			other_hc.CategoryTotoalAmount += hc.CategoryTotoalAmount
			other_hc.HomeSubCategoryList = append(other_hc.HomeSubCategoryList,
				homeSubCategory{
					SubCategoryName:        hc.CategoryName,
					SubCategoryTotalAmount: hc.CategoryTotoalAmount})
		} else {
			hr.HomeCategoryList = append(hr.HomeCategoryList, hc)
		}
	}

	if other_hc.CategoryTotoalAmount != 0 {
		hr.HomeCategoryList = append(hr.HomeCategoryList, *other_hc)
	}

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
	CategoryId                 int                          `json:"category_id"`
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
	TransactionDate   string `json:"transaction_date"`
	PaymentId         *int   `json:"payment_id"`
	PaymentName       string `json:"payment_name"`
}

func GetMonthlyVariableResponse(data *[]model.MonthlyVariableData) *monthlyVariableResponse {
	mvr := &monthlyVariableResponse{}

	mvcl := &[]monthlyVariableCategory{}
	for _, category := range *data {
		if containsVariableCategory(mvcl, &category.CategoryName) {
			continue
		}

		mvc := &monthlyVariableCategory{CategoryId: category.CategoryId, CategoryName: category.CategoryName, CategoryTotoalAmount: category.CategoryTotalAmount}

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
							TransactionAmount: transaction.TransactionAmount,
							TransactionDate:   transaction.TransactionDate.Format("2006-01-02"),
							PaymentId:         &transaction.PaymentId,
							PaymentName:       transaction.PaymentName}

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

func GetTotalSpendingResponse(data *[]model.TotalSpendingData) *totalSpendingResponse {
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

type paymentGroupResponse struct {
	TotalSpending       int           `json:"total_spending"`
	LastMonthTotalSpend int           `json:"last_month_total_spending"`
	MonthOverMonthSum   *float64      `json:"month_over_month_sum"`
	PaymentList         []paymentList `json:"payment_list"`
}

type paymentList struct {
	PaymentId              int                  `json:"payment_id"`
	PaymentName            string               `json:"payment_name"`
	PaymentAmount          int                  `json:"payment_amount"`
	PaymentTypeId          *int                 `json:"payment_type_id"`
	PaymentTypeName        string               `json:"payment_type_name"`
	IsPaymentDueLater      bool                 `json:"is_payment_due_later"`
	LastMonthSum           *int                 `json:"last_month_sum"`
	MonthOverMonth         *float64             `json:"month_over_month"`
	PaymentTransactionList []paymentTransaction `json:"transaction_list"`
}

type paymentTransaction struct {
	TransactionId     int    `json:"transaction_id"`
	TransactionName   string `json:"transaction_name"`
	TransactionAmount int    `json:"transaction_amount"`
	TransactionDate   string `json:"transaction_date"`
	CategoryId        int    `json:"category_id"`
	CategoryName      string `json:"category_name"`
	SubCategoryId     int    `json:"sub_category_id"`
	SubCategoryName   string `json:"sub_category_name"`
	FixedFlg          bool   `json:"fixed_flg"`
}

func GetPaymentGroupResponse(data *[]model.PaymentGroupTransaction, last_month_data *[]model.PaymentGroupTransaction) *paymentGroupResponse {
	pgr := &paymentGroupResponse{PaymentList: []paymentList{}}

	pll := &[]paymentList{}
	for _, payment := range *data {
		if containsPaymentList(pll, &payment.PaymentName) {
			continue
		}

		var paymentTypeId *int
		if payment.PaymentTypeId != 0 {
			paymentTypeId = &payment.PaymentTypeId
		}
		pl := &paymentList{PaymentId: payment.PaymentId, PaymentName: payment.PaymentName, PaymentAmount: payment.PaymentAmount,
			PaymentTypeId: paymentTypeId, PaymentTypeName: payment.PaymentTypeName,
			IsPaymentDueLater: payment.IsPaymentDueLater, LastMonthSum: nil, MonthOverMonth: nil}

		for _, last_payment := range *last_month_data {
			if payment.PaymentId == last_payment.PaymentId {
				pl.LastMonthSum = &last_payment.PaymentAmount
				mom := (float64(payment.PaymentAmount-last_payment.PaymentAmount) * 100) / float64(last_payment.PaymentAmount)
				round_mom := math.Round(mom*100) / 100
				pl.MonthOverMonth = &round_mom
			}
		}

		for _, tran := range *data {
			if pl.PaymentName == tran.PaymentName {
				pl.PaymentTransactionList = append(pl.PaymentTransactionList, paymentTransaction{
					TransactionId:     tran.TransactionId,
					TransactionName:   tran.TransactionName,
					TransactionAmount: tran.TransactionAmount,
					TransactionDate:   tran.TransactionDate.Format("2006-01-02"),
					CategoryId:        tran.CategoryId,
					CategoryName:      tran.CategoryName,
					SubCategoryId:     tran.SubCategoryId,
					SubCategoryName:   tran.SubCategoryName,
					FixedFlg:          tran.FixedFlg})
			}
		}

		*pll = append(*pll, *pl)
		pgr.TotalSpending += payment.PaymentAmount
	}

	// 前月合計の計算
	for _, last_payment := range *last_month_data {
		pgr.LastMonthTotalSpend += last_payment.PaymentAmount
	}

	// 前月比の計算
	if pgr.TotalSpending != 0 && pgr.LastMonthTotalSpend != 0 {
		moms := (float64(pgr.TotalSpending-pgr.LastMonthTotalSpend) * 100) / float64(pgr.LastMonthTotalSpend)
		round_mom := math.Round(moms*100) / 100
		pgr.MonthOverMonthSum = &round_mom
	}

	for i, pl := range *pll {
		if pl.PaymentName == "" {
			(*pll)[i].PaymentName = "未分類"
		}
	}

	pgr.PaymentList = *pll

	return pgr
}

type monthlyWithdrawalAmountResponse struct {
	WithdrawalList []monthlyWithdrawalAmount `json:"withdrawal_list"`
}

type monthlyWithdrawalAmount struct {
	PaymentId            int    `json:"payment_id"`
	PaymentName          string `json:"payment_name"`
	PaymentDate          int    `json:"payment_date"`
	AggregationStartDate string `json:"aggregation_start_date"`
	AggregationEndDate   string `json:"aggregation_end_date"`
	WithdrawalAmount     int    `json:"withdrawal_amount"`
}

func GetMonthlyWithdrawalAmount(data []*model.MonthlyWithdrawalAmountList) *monthlyWithdrawalAmountResponse {
	mwal := &monthlyWithdrawalAmountResponse{WithdrawalList: []monthlyWithdrawalAmount{}}

	for _, item := range data {
		mwal.WithdrawalList = append(mwal.WithdrawalList,
			monthlyWithdrawalAmount{
				PaymentId:            item.PaymentId,
				PaymentName:          item.PaymentName,
				PaymentDate:          item.PaymentDate,
				AggregationStartDate: item.AggregationStartDate,
				AggregationEndDate:   item.AggregationEndDate,
				WithdrawalAmount:     item.WithdrawalAmount})
	}
	return mwal
}

func containsPaymentList(data_list *[]paymentList, payment_name *string) bool {
	for _, v := range *data_list {
		if v.PaymentName == *payment_name {
			return true
		}
	}
	return false
}

type frequentTransactionResponse struct {
	FrequentTransactionlist []frequentTransaction `json:"transaction_list"`
}

type frequentTransaction struct {
	TransactionName string `json:"transaction_name"`
	CategoryId      int    `json:"category_id"`
	SubCategoryId   int    `json:"sub_category_id"`
	FixedFlg        bool   `json:"fixed_flg"`
	PaymentId       *int   `json:"payment_id"`
	CategoryName    string `json:"category_name"`
	SubCategoryName string `json:"sub_category_name"`
}

func containsFrequentName(transactions []frequentTransaction, transactionName string) bool {
	for _, transaction := range transactions {
		if transaction.TransactionName == transactionName {
			return true
		}
	}
	return false
}

func GetFrequentTransactionResponse(data *[]model.FrequentTransactionName) *frequentTransactionResponse {
	ftr := &frequentTransactionResponse{FrequentTransactionlist: []frequentTransaction{}}

	for _, tran := range *data {

		var paymentId *int
		if tran.PaymentId != 0 {
			paymentId = &tran.PaymentId
		}
		if exist := containsFrequentName(ftr.FrequentTransactionlist, tran.TransactionName); !exist {
			ftr.FrequentTransactionlist = append(ftr.FrequentTransactionlist, frequentTransaction{
				TransactionName: tran.TransactionName,
				CategoryId:      tran.CategoryId,
				SubCategoryId:   tran.SubCategoryId,
				FixedFlg:        tran.FixedFlg,
				PaymentId:       paymentId,
				CategoryName:    tran.CategoryName,
				SubCategoryName: tran.SubCategoryName,
			})
		}
	}

	return ftr
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
