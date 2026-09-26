package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

const (
	schedulerAudienceEnvironmentKey            = "SCHEDULER_AUDIENCE"
	schedulerServiceAccountEmailEnvironmentKey = "SCHEDULER_SERVICE_ACCOUNT_EMAIL"
	schedulerJobNameEnvironmentKey             = "JOB_NAME"
	googleAccountsIssuer                       = "accounts.google.com"
	googleAccountsHTTPSIssuer                  = "https://accounts.google.com"
)

type SchedulerAuthConfig struct {
	Audience            string
	ServiceAccountEmail string
	JobName             string
}

type SchedulerIDTokenValidator interface {
	Validate(context.Context, string, string) (*idtoken.Payload, error)
}

type GoogleSchedulerIDTokenValidator struct{}

func (GoogleSchedulerIDTokenValidator) Validate(ctx context.Context, token string, audience string) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, token, audience)
}

func SchedulerAuthConfigFromEnvironment() (SchedulerAuthConfig, error) {
	config := SchedulerAuthConfig{
		Audience:            strings.TrimSpace(os.Getenv(schedulerAudienceEnvironmentKey)),
		ServiceAccountEmail: strings.TrimSpace(os.Getenv(schedulerServiceAccountEmailEnvironmentKey)),
		JobName:             strings.TrimSpace(os.Getenv(schedulerJobNameEnvironmentKey)),
	}

	for key, value := range map[string]string{
		schedulerAudienceEnvironmentKey:            config.Audience,
		schedulerServiceAccountEmailEnvironmentKey: config.ServiceAccountEmail,
		schedulerJobNameEnvironmentKey:             config.JobName,
	} {
		if value == "" {
			return SchedulerAuthConfig{}, fmt.Errorf("%s must be set", key)
		}
	}

	return config, nil
}

func SchedulerAuthMiddleware(validator SchedulerIDTokenValidator, config SchedulerAuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString, ok := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if !ok {
				return respondAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "A valid Scheduler Bearer token is required")
			}
			if validator == nil || config.Audience == "" || config.ServiceAccountEmail == "" {
				log.Printf("event=scheduler_authentication_error type=configuration_unavailable")
				return respondAuthError(c, http.StatusInternalServerError, "AUTH_SERVICE_UNAVAILABLE", "Scheduler authentication is unavailable")
			}

			payload, err := validator.Validate(c.Request().Context(), tokenString, config.Audience)
			if err != nil {
				log.Printf("event=scheduler_authentication_rejected type=invalid_token error=%q", err)
				return respondAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired Scheduler ID token")
			}
			if payload == nil || !isGoogleAccountsIssuer(payload.Issuer) {
				log.Printf("event=scheduler_authentication_rejected type=invalid_issuer")
				return respondAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired Scheduler ID token")
			}

			email, _ := payload.Claims["email"].(string)
			emailVerified, _ := payload.Claims["email_verified"].(bool)
			if !emailVerified || email != config.ServiceAccountEmail {
				log.Printf("event=scheduler_authentication_rejected type=identity_not_allowed email=%q", email)
				return respondAuthError(c, http.StatusForbidden, "FORBIDDEN", "Scheduler identity is not allowed")
			}

			return next(c)
		}
	}
}

func isGoogleAccountsIssuer(issuer string) bool {
	return issuer == googleAccountsIssuer || issuer == googleAccountsHTTPSIssuer
}
