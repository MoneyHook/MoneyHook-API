package handler

import (
	common "MoneyHook/MoneyHook-API/common"
	"MoneyHook/MoneyHook-API/handler/request"
	"MoneyHook/MoneyHook-API/model"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (h *Handler) googleSignIn(c echo.Context) error {
	var googleSignIn model.GoogleSignIn

	req := &request.GoogleSignInRequest{}
	if err := req.Bind(c, &googleSignIn); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	// 受け取ったユーザーIDがuserテーブルに存在するかどうかチェック
	if result := h.userStore.UserExists(&googleSignIn.UserId); *result != "" {
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

func (h *Handler) GetUserId(c echo.Context) (string, error) {
	// Authorizationヘッダからトークンを抽出
	token := c.Request().Header.Get(echo.HeaderAuthorization)

	var userNo string

	if EnableFirebaseAuth() {
		user, err := h.firebaseClient.VerifyIDToken(context.Background(), token)
		if err != nil {
			log.Printf("GetUserId FirebaseAuth error: %v\n", err)
			return "", err
		}
		email := user.Claims["email"]

		user_id := convHash(email.(string))

		result, err := h.userStore.ExtractUserNoFromUserId(&user_id)
		if err != nil {
			log.Printf("GetUserId extract user error: %v\n", err)
			return "", err
		}
		userNo = *result
	} else {
		// トークンからUserNoを抽出(DBのハッシュかされたIDトークンを見る方法)
		result, err := h.userStore.ExtractUserNoFromToken(&token)
		if err != nil {
			log.Printf("GetUserId  extract user error: %v\n", err)
			return "", err
		}
		userNo = *result
	}

	return userNo, nil
}

/* 環境変数からFirebaseAuthを行うかどうかを取得 */
func EnableFirebaseAuth() bool {

	fa := common.GetEnv("ENABLE_FIREBASE_AUTH", "true")

	result, _ := strconv.ParseBool(fa)

	return result
}

/* 受け取った文字列をハッシュ化 */
func convHash(message string) string {
	h := sha256.New()
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
