package handler

import (
	"MoneyHook/MoneyHook-API/model"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) googleSignIn(c echo.Context) error {
	var googleSignIn model.GoogleSignIn

	req := &GoogleSignInRequest{}
	if err := req.bind(c, &googleSignIn); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	// 受け取ったユーザーIDがuserテーブルに存在するかどうかチェック
	if result := h.userStore.UserExists(&googleSignIn.UserId); *result != 0 {
		// Yes: user_tokenテーブルのtokenを更新
		googleSignIn.UserNo = *result
		h.userStore.UpdateToken(&googleSignIn)
	} else {
		// No: userテーブル・user_tokenテーブルに登録
		result := h.userStore.CreateUser(&googleSignIn)
		fmt.Println(result.UserNo)

		h.userStore.CreateToken(&googleSignIn)
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
