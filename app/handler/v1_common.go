package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type v1ErrorResponse struct {
	Status      string            `json:"status"`
	Code        string            `json:"code"`
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}

func decodeV1JSON(c echo.Context, destination any) error {
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body contains multiple JSON values")
		}
		return err
	}
	return nil
}

func respondV1Error(c echo.Context, status int, code string, message string, fieldErrors map[string]string) error {
	return c.JSON(status, v1ErrorResponse{
		Status:      "error",
		Code:        code,
		Message:     message,
		FieldErrors: fieldErrors,
	})
}

func respondV1Unauthorized(c echo.Context) error {
	return respondV1Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "認証に失敗しました", nil)
}

func (h *Handler) GetV1UserId(c echo.Context) (string, error) {
	return h.GetUserId(c)
}
