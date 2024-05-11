package handler

import (
	"MoneyHook/MoneyHook-API/model"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"

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
		h.userStore.CreateUser(&googleSignIn)

		h.userStore.CreateToken(&googleSignIn)
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) GetUserId(c echo.Context) int {
	// Authorizationヘッダからトークンを抽出
	token := c.Request().Header["Authorization"][0]
	var userNo int

	if EnableFirebaseAuth() {
		user, err := h.firebaseClient.VerifyIDToken(context.Background(), token)
		if err != nil {
			fmt.Println("firebase auth error!!")
			return 3
		}
		email := user.Claims["email"]

		user_id := convHash(email.(string))

		userNo = *h.userStore.ExtractUserNoFromUserId(&user_id)
	} else {
		// トークンからUserNoを抽出(DBのハッシュかされたIDトークンを見る方法)
		userNo = *h.userStore.ExtractUserNoFromToken(&token)
	}

	return userNo
}

/* 環境変数からFirebaseAuthを行うかどうかを取得 */
func EnableFirebaseAuth() bool {

	fa := os.Getenv("ENABLE_FIREBASE_AUTH")
	if len(fa) == 0 {
		// 設定されていない場合、True
		return true
	}

	result, _ := strconv.ParseBool(fa)

	return result
}

/* 受け取った文字列をハッシュ化 */
func convHash(message string) string {
	h := sha256.New()
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
