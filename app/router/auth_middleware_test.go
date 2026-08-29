package router

import (
	userdomain "MoneyHook/MoneyHook-API/user"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
)

type fakeTokenVerifier struct {
	token *auth.Token
	err   error
}

func (f fakeTokenVerifier) VerifyIDToken(context.Context, string) (*auth.Token, error) {
	return f.token, f.err
}

type fakeUserStore struct {
	userNo       string
	err          error
	firebaseUID  string
	legacyUserID string
}

func (f *fakeUserStore) ResolveFirebaseUser(firebaseUID string, legacyUserID string) (string, error) {
	f.firebaseUID = firebaseUID
	f.legacyUserID = legacyUserID
	return f.userNo, f.err
}

func googleToken() *auth.Token {
	return &auth.Token{
		UID:      "firebase-user",
		Firebase: auth.FirebaseInfo{SignInProvider: googleSignInProvider},
		Claims: map[string]interface{}{
			"email":          "person@example.com",
			"email_verified": true,
		},
	}
}

func executeMiddleware(t *testing.T, header string, verifier IDTokenVerifier, store userdomain.Store) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	if header != "" {
		req.Header.Set(echo.HeaderAuthorization, header)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := FirebaseAuthMiddleware(verifier, store)(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"uid":     GetFirebaseUID(c),
			"user_no": GetUserNo(c),
		})
	})
	if err := handler(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
	return rec
}

func TestFirebaseAuthMiddlewareRequiresBearerToken(t *testing.T) {
	for _, header := range []string{"", "Basic abc", "Bearer", "Bearer a b"} {
		rec := executeMiddleware(t, header, nil, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("header=%q status=%d, want %d", header, rec.Code, http.StatusUnauthorized)
		}
	}
}

func TestFirebaseAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	rec := executeMiddleware(t, "Bearer invalid", fakeTokenVerifier{err: errors.New("invalid")}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestFirebaseAuthMiddlewareRejectsEmptyVerifiedToken(t *testing.T) {
	rec := executeMiddleware(t, "Bearer empty", fakeTokenVerifier{}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestFirebaseAuthMiddlewareRequiresGoogleProvider(t *testing.T) {
	token := googleToken()
	token.Firebase.SignInProvider = "password"
	rec := executeMiddleware(t, "Bearer password-token", fakeTokenVerifier{token: token}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestFirebaseAuthMiddlewareRequiresVerifiedEmail(t *testing.T) {
	token := googleToken()
	token.Claims["email_verified"] = false
	rec := executeMiddleware(t, "Bearer google-token", fakeTokenVerifier{token: token}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestFirebaseAuthMiddlewareResolvesVerifiedGoogleUser(t *testing.T) {
	store := &fakeUserStore{userNo: "42"}
	rec := executeMiddleware(t, "bearer google-token", fakeTokenVerifier{token: googleToken()}, store)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if store.firebaseUID != "firebase-user" {
		t.Fatalf("firebase uid=%q", store.firebaseUID)
	}
	if store.legacyUserID != "542d240129883c019e106e3b1b2d3f3cb3537c43c425364de8e951d5a3083345" {
		t.Fatalf("legacy user id=%q", store.legacyUserID)
	}
}

func TestFirebaseAuthMiddlewareReportsIdentityConflict(t *testing.T) {
	store := &fakeUserStore{err: userdomain.ErrIdentityConflict}
	rec := executeMiddleware(t, "Bearer google-token", fakeTokenVerifier{token: googleToken()}, store)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusConflict)
	}
}
