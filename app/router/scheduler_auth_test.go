package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

type schedulerIDTokenValidatorStub struct {
	payload          *idtoken.Payload
	err              error
	receivedToken    string
	receivedAudience string
}

func (s *schedulerIDTokenValidatorStub) Validate(_ context.Context, token string, audience string) (*idtoken.Payload, error) {
	s.receivedToken = token
	s.receivedAudience = audience
	return s.payload, s.err
}

func executeSchedulerMiddleware(
	t *testing.T,
	header string,
	validator SchedulerIDTokenValidator,
	config SchedulerAuthConfig,
) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	request := httptest.NewRequest(http.MethodPost, "/api/job/daily", nil)
	request.Header.Set(echo.HeaderAuthorization, header)
	recorder := httptest.NewRecorder()
	handler := SchedulerAuthMiddleware(validator, config)(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if err := handler(e.NewContext(request, recorder)); err != nil {
		t.Fatal(err)
	}
	return recorder
}

func validSchedulerAuthConfig() SchedulerAuthConfig {
	return SchedulerAuthConfig{
		Audience:            "https://api.example.com/api/job/daily",
		ServiceAccountEmail: "scheduler@example-project.iam.gserviceaccount.com",
		JobName:             "daily",
	}
}

func validSchedulerPayload() *idtoken.Payload {
	return &idtoken.Payload{
		Issuer: "https://accounts.google.com",
		Claims: map[string]interface{}{
			"email":          "scheduler@example-project.iam.gserviceaccount.com",
			"email_verified": true,
		},
	}
}

func TestSchedulerAuthConfigFromEnvironment(t *testing.T) {
	t.Setenv(schedulerAudienceEnvironmentKey, " https://api.example.com/api/job/daily ")
	t.Setenv(schedulerServiceAccountEmailEnvironmentKey, " scheduler@example-project.iam.gserviceaccount.com ")
	t.Setenv(schedulerJobNameEnvironmentKey, " daily ")

	config, err := SchedulerAuthConfigFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if config != validSchedulerAuthConfig() {
		t.Fatalf("config = %+v", config)
	}
}

func TestSchedulerAuthConfigRequiresEveryValue(t *testing.T) {
	for _, missing := range []string{
		schedulerAudienceEnvironmentKey,
		schedulerServiceAccountEmailEnvironmentKey,
		schedulerJobNameEnvironmentKey,
	} {
		t.Run(missing, func(t *testing.T) {
			t.Setenv(schedulerAudienceEnvironmentKey, "https://api.example.com/api/job/daily")
			t.Setenv(schedulerServiceAccountEmailEnvironmentKey, "scheduler@example-project.iam.gserviceaccount.com")
			t.Setenv(schedulerJobNameEnvironmentKey, "daily")
			t.Setenv(missing, " ")

			if _, err := SchedulerAuthConfigFromEnvironment(); err == nil {
				t.Fatal("expected missing configuration to fail")
			}
		})
	}
}

func TestSchedulerAuthMiddlewareRejectsMissingOrMalformedBearerToken(t *testing.T) {
	for _, header := range []string{"", "Basic token", "Bearer", "Bearer one two"} {
		recorder := executeSchedulerMiddleware(t, header, &schedulerIDTokenValidatorStub{}, validSchedulerAuthConfig())
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("header %q returned %d", header, recorder.Code)
		}
	}
}

func TestSchedulerAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	validator := &schedulerIDTokenValidatorStub{err: errors.New("invalid signature")}
	recorder := executeSchedulerMiddleware(t, "Bearer scheduler-token", validator, validSchedulerAuthConfig())
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", recorder.Code)
	}
	if validator.receivedAudience != validSchedulerAuthConfig().Audience {
		t.Fatalf("audience = %q", validator.receivedAudience)
	}
}

func TestSchedulerAuthMiddlewareRejectsUnexpectedIssuer(t *testing.T) {
	payload := validSchedulerPayload()
	payload.Issuer = "https://cloud.google.com/iap"
	recorder := executeSchedulerMiddleware(t, "Bearer scheduler-token", &schedulerIDTokenValidatorStub{payload: payload}, validSchedulerAuthConfig())
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestSchedulerAuthMiddlewareRejectsUnapprovedIdentity(t *testing.T) {
	for _, mutate := range []func(*idtoken.Payload){
		func(payload *idtoken.Payload) {
			payload.Claims["email"] = "other@example-project.iam.gserviceaccount.com"
		},
		func(payload *idtoken.Payload) { payload.Claims["email_verified"] = false },
	} {
		payload := validSchedulerPayload()
		mutate(payload)
		recorder := executeSchedulerMiddleware(t, "Bearer scheduler-token", &schedulerIDTokenValidatorStub{payload: payload}, validSchedulerAuthConfig())
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d", recorder.Code)
		}
	}
}

func TestSchedulerAuthMiddlewareAllowsApprovedIdentity(t *testing.T) {
	validator := &schedulerIDTokenValidatorStub{payload: validSchedulerPayload()}
	recorder := executeSchedulerMiddleware(t, "Bearer scheduler-token", validator, validSchedulerAuthConfig())
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d", recorder.Code)
	}
	if validator.receivedToken != "scheduler-token" {
		t.Fatalf("token = %q", validator.receivedToken)
	}
}
