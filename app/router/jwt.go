package router

import (
	common "MoneyHook/MoneyHook-API/common"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

/*
GenerateJWTは、JWTトークンを生成し、クライアントに返します。
クライアントの認証情報を元にJWTトークンを生成し、HTTPレスポンスとして返します。
JWTの秘密鍵は環境変数から取得し、デフォルト値は "secret" です。

パラメータ:
  - c: echo.Context - Echoフレームワークのコンテキスト

戻り値:
  - error: エラーが発生した場合に返されます。成功した場合はnil。
*/
func GenerateJWT(c echo.Context) error {
	// JWTの秘密鍵
	var jwtSecret = []byte(common.GetEnv("JWT_SECRET", "secret"))

	// JWTトークン生成
	claims := jwt.MapClaims{
		"authorized":      true,
		"application_key": common.GetEnv("APPLICATION_KEY", "thisIsApplicationKey"),
		"exp":             time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Could not generate token")
	}

	cookie := &http.Cookie{
		Name:  "ApplicationKey",
		Value: jwtToken,
	}
	c.SetCookie(cookie)

	return c.String(http.StatusOK, "Success, SetCookie")
}

/*
ValidateJwtは、JWTトークンを検証するためのミドルウェア設定を返します。
トークンは "ApplicationKey" Cookieから取得され、特定の条件を満たす場合にスキップされます。
JWTの秘密鍵は環境変数から取得し、デフォルト値は "secret" です。

戻り値:
  - middleware.KeyAuthConfig: JWTトークンの検証設定を含む構造体。
*/
func ValidateJwt() middleware.KeyAuthConfig {
	return middleware.KeyAuthConfig{
		KeyLookup: "cookie:ApplicationKey",
		Skipper:   isIgnoreList,
		Validator: func(key string, c echo.Context) (bool, error) {
			token, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method: %v", token.Header["alg"])
				}

				// JWTの秘密鍵
				var jwtSecret = []byte(common.GetEnv("JWT_SECRET", "secret"))

				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				return false, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if claims["application_key"] == common.GetEnv("APPLICATION_KEY", "thisIsApplicationKey") {
					return true, nil
				} else {
					return false, echo.NewHTTPError(http.StatusUnauthorized, "Missing value claim")
				}
			} else {
				return false, echo.NewHTTPError(http.StatusUnauthorized, "Missing value claim")
			}
		},
	}
}

func isIgnoreList(c echo.Context) bool {
	for _, igonoreValue := range common.IgnoreVerifyApiKeyList {
		if c.Request().RequestURI == igonoreValue {
			return true
		}
	}
	return false
}
