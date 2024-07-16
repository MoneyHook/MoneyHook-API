package request

import (
	"MoneyHook/MoneyHook-API/model"

	"github.com/labstack/echo/v4"
)

type GoogleSignInRequest struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

func (r *GoogleSignInRequest) Bind(c echo.Context, u *model.GoogleSignIn) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.UserId = r.UserId
	u.Token = r.Token

	return nil
}
