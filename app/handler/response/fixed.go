package response

import "MoneyHook/MoneyHook-API/model"

type fixedResponse struct {
	MonthlyTransactionList []fixedResponseData `json:"monthly_transaction_list"`
}

type fixedResponseData struct {
	MonthlyTransactionId     string  `json:"monthly_transaction_id"`
	MonthlyTransactionName   string  `json:"monthly_transaction_name"`
	MonthlyTransactionAmount int     `json:"monthly_transaction_amount"`
	MonthlyTransactionSign   int     `json:"monthly_transaction_sign"`
	MonthlyTransactionDate   int     `json:"monthly_transaction_date"`
	CategoryId               string  `json:"category_id"`
	CategoryName             string  `json:"category_name"`
	SubCategoryId            string  `json:"sub_category_id"`
	SubCategoryName          string  `json:"sub_category_name"`
	PaymentId                *string `json:"payment_id"`
}

func GetFixedResponse(data_list *[]model.GetFixed) *fixedResponse {
	fr := &[]fixedResponseData{}

	for _, data := range *data_list {

		var paymentId *string
		if data.PaymentId != "" {
			paymentId = &data.PaymentId
		}
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
				PaymentId:                paymentId,
			})
	}

	return &fixedResponse{MonthlyTransactionList: *fr}
}

type deletedFixedResponse struct {
	MonthlyTransactionId     string `json:"monthly_transaction_id"`
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
