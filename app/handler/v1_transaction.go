package handler

import (
	"MoneyHook/MoneyHook-API/model"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
)

type v1TransactionInput struct {
	TransactionDate string  `json:"transaction_date"`
	TransactionTime *string `json:"transaction_time"`
	TransactionName string  `json:"transaction_name"`
	Amount          int64   `json:"amount"`
	Sign            int     `json:"sign"`
	CategoryId      string  `json:"category_id"`
	SubCategoryId   string  `json:"sub_category_id"`
	FixedFlg        *bool   `json:"fixed_flg"`
	PaymentId       *string `json:"payment_id"`
}

type v1TransactionRequest struct {
	Transaction v1TransactionInput `json:"transaction"`
}

type v1TransactionResource struct {
	TransactionId   string  `json:"transaction_id"`
	TransactionDate string  `json:"transaction_date"`
	TransactionTime *string `json:"transaction_time"`
	TransactionName string  `json:"transaction_name"`
	Amount          int64   `json:"amount"`
	Sign            int     `json:"sign"`
	SignedAmount    int64   `json:"signed_amount"`
	CategoryId      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	SubCategoryId   string  `json:"sub_category_id"`
	SubCategoryName string  `json:"sub_category_name"`
	FixedFlg        bool    `json:"fixed_flg"`
	PaymentId       *string `json:"payment_id"`
	PaymentName     *string `json:"payment_name"`
}

type v1TransactionResponse struct {
	Transaction v1TransactionResource `json:"transaction"`
}

type v1TransactionUpdateResponse struct {
	Transaction             v1TransactionResource `json:"transaction"`
	PreviousTransactionDate string                `json:"previous_transaction_date"`
}

func (h *Handler) getV1Transaction(c echo.Context) error {
	userId, err := h.GetV1UserId(c)
	if err != nil {
		return respondV1Unauthorized(c)
	}
	transactionId := c.Param("transactionId")
	if !isPositiveNumericID(transactionId) {
		return respondV1Error(c, http.StatusBadRequest, "INVALID_PATH_PARAMETER", "取引IDが不正です", nil)
	}
	result, err := h.transactionStore.GetV1Transaction(userId, transactionId)
	if err != nil {
		return h.respondV1TransactionStoreError(c, err)
	}
	return c.JSON(http.StatusOK, v1TransactionResponse{Transaction: newV1TransactionResource(result)})
}

func (h *Handler) createV1Transaction(c echo.Context) error {
	userId, err := h.GetV1UserId(c)
	if err != nil {
		return respondV1Unauthorized(c)
	}
	var request v1TransactionRequest
	if err := decodeV1JSON(c, &request); err != nil {
		return respondV1Error(c, http.StatusBadRequest, "INVALID_JSON", "JSON形式が不正です", nil)
	}
	fieldErrors := validateV1TransactionInput(request.Transaction)
	if len(fieldErrors) > 0 {
		return respondV1Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "入力内容を確認してください", fieldErrors)
	}
	input := request.Transaction.toModel(userId, "")
	result, err := h.transactionStore.CreateV1Transaction(&input)
	if err != nil {
		return h.respondV1TransactionStoreError(c, err)
	}
	return c.JSON(http.StatusCreated, v1TransactionResponse{Transaction: newV1TransactionResource(result)})
}

func (h *Handler) updateV1Transaction(c echo.Context) error {
	userId, err := h.GetV1UserId(c)
	if err != nil {
		return respondV1Unauthorized(c)
	}
	transactionId := c.Param("transactionId")
	if !isPositiveNumericID(transactionId) {
		return respondV1Error(c, http.StatusBadRequest, "INVALID_PATH_PARAMETER", "取引IDが不正です", nil)
	}
	var request v1TransactionRequest
	if err := decodeV1JSON(c, &request); err != nil {
		return respondV1Error(c, http.StatusBadRequest, "INVALID_JSON", "JSON形式が不正です", nil)
	}
	fieldErrors := validateV1TransactionInput(request.Transaction)
	if len(fieldErrors) > 0 {
		return respondV1Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "入力内容を確認してください", fieldErrors)
	}
	input := request.Transaction.toModel(userId, transactionId)
	result, previousDate, err := h.transactionStore.UpdateV1Transaction(&input)
	if err != nil {
		return h.respondV1TransactionStoreError(c, err)
	}
	return c.JSON(http.StatusOK, v1TransactionUpdateResponse{
		Transaction:             newV1TransactionResource(result),
		PreviousTransactionDate: previousDate,
	})
}

func (h *Handler) deleteV1Transaction(c echo.Context) error {
	userId, err := h.GetV1UserId(c)
	if err != nil {
		return respondV1Unauthorized(c)
	}
	transactionId := c.Param("transactionId")
	if !isPositiveNumericID(transactionId) {
		return respondV1Error(c, http.StatusBadRequest, "INVALID_PATH_PARAMETER", "取引IDが不正です", nil)
	}
	if err := h.transactionStore.DeleteV1Transaction(userId, transactionId); err != nil {
		return h.respondV1TransactionStoreError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) respondV1TransactionStoreError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, transactiondomain.ErrNotFound):
		return respondV1Error(c, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "取引が見つかりません", nil)
	case errors.Is(err, transactiondomain.ErrInvalidRelation):
		return respondV1Error(c, http.StatusNotFound, "RELATED_RESOURCE_NOT_FOUND", "利用可能なカテゴリ、サブカテゴリ、または支払方法が見つかりません", nil)
	default:
		return respondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーが発生しました", nil)
	}
}

func (input v1TransactionInput) toModel(userId string, transactionId string) model.V1TransactionWrite {
	fixedFlg := false
	if input.FixedFlg != nil {
		fixedFlg = *input.FixedFlg
	}
	return model.V1TransactionWrite{
		TransactionId:   transactionId,
		UserId:          userId,
		TransactionDate: input.TransactionDate,
		TransactionTime: input.TransactionTime,
		TransactionName: strings.TrimSpace(input.TransactionName),
		Amount:          input.Amount,
		Sign:            input.Sign,
		CategoryId:      input.CategoryId,
		SubCategoryId:   input.SubCategoryId,
		FixedFlg:        fixedFlg,
		PaymentId:       input.PaymentId,
	}
}

func newV1TransactionResource(transaction *model.V1Transaction) v1TransactionResource {
	return v1TransactionResource{
		TransactionId:   transaction.TransactionId,
		TransactionDate: transaction.TransactionDate,
		TransactionTime: transaction.TransactionTime,
		TransactionName: transaction.TransactionName,
		Amount:          transaction.Amount,
		Sign:            transaction.Sign,
		SignedAmount:    transaction.SignedAmount,
		CategoryId:      transaction.CategoryId,
		CategoryName:    transaction.CategoryName,
		SubCategoryId:   transaction.SubCategoryId,
		SubCategoryName: transaction.SubCategoryName,
		FixedFlg:        transaction.FixedFlg,
		PaymentId:       transaction.PaymentId,
		PaymentName:     transaction.PaymentName,
	}
}

func validateV1TransactionInput(input v1TransactionInput) map[string]string {
	errors := map[string]string{}
	name := strings.TrimSpace(input.TransactionName)
	nameLength := utf8.RuneCountInString(name)
	if nameLength == 0 || nameLength > 32 {
		errors["transaction.transaction_name"] = "1〜32文字で入力してください"
	}
	if input.Amount < 1 || input.Amount > 9_999_999 {
		errors["transaction.amount"] = "1〜9,999,999で入力してください"
	}
	if input.Sign != -1 && input.Sign != 1 {
		errors["transaction.sign"] = "-1または1を指定してください"
	}
	if !isISODate(input.TransactionDate) {
		errors["transaction.transaction_date"] = "YYYY-MM-DD形式の実在する日付を指定してください"
	}
	if input.TransactionTime != nil && !isLocalTime(*input.TransactionTime) {
		errors["transaction.transaction_time"] = "HH:mm形式の実在する時刻を指定してください"
	}
	if !isPositiveNumericID(input.CategoryId) {
		errors["transaction.category_id"] = "正の数値IDを指定してください"
	}
	if !isPositiveNumericID(input.SubCategoryId) {
		errors["transaction.sub_category_id"] = "正の数値IDを指定してください"
	}
	if input.FixedFlg == nil {
		errors["transaction.fixed_flg"] = "trueまたはfalseを指定してください"
	}
	if input.PaymentId != nil && !isPositiveNumericID(*input.PaymentId) {
		errors["transaction.payment_id"] = "正の数値IDまたはnullを指定してください"
	}
	return errors
}

func isISODate(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

func isLocalTime(value string) bool {
	parsed, err := time.Parse("15:04", value)
	return err == nil && parsed.Format("15:04") == value
}

func isPositiveNumericID(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err == nil && parsed > 0
}
