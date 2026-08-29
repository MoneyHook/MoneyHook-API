package router

import (
	userdomain "MoneyHook/MoneyHook-API/user"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyFirebaseUID = "firebase_uid"
	ContextKeyUserNo      = "user_no"
	googleSignInProvider  = "google.com"
)

type IDTokenVerifier interface {
	VerifyIDToken(context.Context, string) (*auth.Token, error)
}

func FirebaseAuthMiddleware(verifier IDTokenVerifier, userStore userdomain.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString, ok := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if !ok {
				return respondAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "A valid Bearer token is required")
			}
			if verifier == nil {
				log.Printf("event=authentication_error type=verifier_unavailable")
				return respondAuthError(c, http.StatusInternalServerError, "AUTH_SERVICE_UNAVAILABLE", "Authentication service is unavailable")
			}

			token, err := verifier.VerifyIDToken(c.Request().Context(), tokenString)
			if err != nil {
				log.Printf("event=authentication_rejected type=invalid_token error=%q", err)
				return respondAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired Firebase ID token")
			}
			if token == nil || token.UID == "" || token.Firebase.SignInProvider != googleSignInProvider {
				return respondAuthError(c, http.StatusUnauthorized, "GOOGLE_SIGN_IN_REQUIRED", "Google sign-in is required")
			}
			email, emailVerified := verifiedEmail(token.Claims)
			if email == "" || !emailVerified {
				return respondAuthError(c, http.StatusUnauthorized, "VERIFIED_GOOGLE_EMAIL_REQUIRED", "A verified Google email is required")
			}
			if userStore == nil {
				log.Printf("event=authentication_error type=user_store_unavailable")
				return respondAuthError(c, http.StatusInternalServerError, "AUTH_SERVICE_UNAVAILABLE", "Authentication service is unavailable")
			}

			legacyDigest := sha256.Sum256([]byte(email))
			legacyUserID := hex.EncodeToString(legacyDigest[:])
			userNo, err := userStore.ResolveFirebaseUser(token.UID, legacyUserID)
			if errors.Is(err, userdomain.ErrIdentityConflict) {
				log.Printf("event=authentication_rejected type=identity_conflict firebase_uid=%q", token.UID)
				return respondAuthError(c, http.StatusConflict, "AUTH_IDENTITY_CONFLICT", "The Google identity conflicts with an existing user")
			}
			if err != nil || userNo == "" {
				log.Printf("event=authentication_error type=user_resolution firebase_uid=%q error=%q", token.UID, err)
				return respondAuthError(c, http.StatusInternalServerError, "AUTH_USER_RESOLUTION_FAILED", "The authenticated user could not be resolved")
			}

			c.Set(ContextKeyFirebaseUID, token.UID)
			c.Set(ContextKeyUserNo, userNo)
			ctx := context.WithValue(c.Request().Context(), ContextKeyFirebaseUID, token.UID)
			ctx = context.WithValue(ctx, ContextKeyUserNo, userNo)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func verifiedEmail(claims map[string]interface{}) (string, bool) {
	email, _ := claims["email"].(string)
	verified, _ := claims["email_verified"].(bool)
	return strings.TrimSpace(email), verified
}

func respondAuthError(c echo.Context, status int, code string, message string) error {
	return c.JSON(status, map[string]interface{}{
		"status":  "error",
		"code":    code,
		"message": message,
	})
}

func GetFirebaseUID(c echo.Context) string {
	if uid, ok := c.Get(ContextKeyFirebaseUID).(string); ok && uid != "" {
		return uid
	}
	if uid, ok := c.Request().Context().Value(ContextKeyFirebaseUID).(string); ok && uid != "" {
		return uid
	}
	return ""
}

func GetUserNo(c echo.Context) string {
	if userNo, ok := c.Get(ContextKeyUserNo).(string); ok && userNo != "" {
		return userNo
	}
	if userNo, ok := c.Request().Context().Value(ContextKeyUserNo).(string); ok && userNo != "" {
		return userNo
	}
	return ""
}
