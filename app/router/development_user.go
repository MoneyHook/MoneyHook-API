package router

import (
	common "MoneyHook/MoneyHook-API/common"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"firebase.google.com/go/v4/auth"
)

const developmentUserEnabledEnvironmentKey = "ENABLE_DEVELOPMENT_USER"

type authAdminClient interface {
	GetUser(context.Context, string) (*auth.UserRecord, error)
	ImportUsers(context.Context, []*auth.UserToImport, ...auth.UserImportOption) (*auth.UserImportResult, error)
	UpdateUser(context.Context, string, *auth.UserToUpdate) (*auth.UserRecord, error)
}

var _ authAdminClient = (*auth.Client)(nil)

func DevelopmentUserEnabledFromEnvironment() (bool, error) {
	value, exists := os.LookupEnv(developmentUserEnabledEnvironmentKey)
	return parseDevelopmentUserEnabled(value, exists)
}

func parseDevelopmentUserEnabled(value string, exists bool) (bool, error) {
	value = strings.TrimSpace(value)
	if !exists || value == "" {
		return false, nil
	}

	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s value %q: expected a boolean accepted by strconv.ParseBool", developmentUserEnabledEnvironmentKey, value)
	}
	return enabled, nil
}

func EnsureDevelopmentUser(ctx context.Context, client authAdminClient) error {
	return ensureDevelopmentUser(ctx, client, auth.IsUserNotFound)
}

func ensureDevelopmentUser(
	ctx context.Context,
	client authAdminClient,
	isUserNotFound func(error) bool,
) error {
	existing, err := client.GetUser(ctx, common.DevelopmentUserID)
	if err == nil {
		return reconcileDevelopmentUser(ctx, client, existing)
	}
	if !isUserNotFound(err) {
		return fmt.Errorf("lookup development Firebase user: %w", err)
	}

	result, err := client.ImportUsers(ctx, []*auth.UserToImport{developmentUserImport()})
	if err != nil {
		return fmt.Errorf("create development Firebase user: %w", err)
	}
	if result == nil || result.SuccessCount != 1 || result.FailureCount != 0 {
		return fmt.Errorf("create development Firebase user returned an incomplete result: %+v", result)
	}

	log.Printf(
		"event=development_auth_user action=create uid=%q display_name=%q provider=%q",
		common.DevelopmentUserID,
		common.DevelopmentUserName,
		"google.com",
	)
	return nil
}

func developmentUserImport() *auth.UserToImport {
	return (&auth.UserToImport{}).
		UID(common.DevelopmentUserID).
		Email(common.DevelopmentUserEmail).
		DisplayName(common.DevelopmentUserName).
		EmailVerified(true).
		ProviderData([]*auth.UserProvider{
			{
				UID:         common.DevelopmentUserID,
				ProviderID:  "google.com",
				Email:       common.DevelopmentUserEmail,
				DisplayName: common.DevelopmentUserName,
			},
		})
}

func reconcileDevelopmentUser(ctx context.Context, client authAdminClient, existing *auth.UserRecord) error {
	if existing == nil || existing.UserInfo == nil {
		return fmt.Errorf("development Firebase user %q has no user information", common.DevelopmentUserID)
	}

	update := &auth.UserToUpdate{}
	needsUpdate := false
	if existing.DisplayName != common.DevelopmentUserName {
		update.DisplayName(common.DevelopmentUserName)
		needsUpdate = true
	}
	if existing.Email != common.DevelopmentUserEmail {
		update.Email(common.DevelopmentUserEmail)
		needsUpdate = true
	}
	if !existing.EmailVerified {
		update.EmailVerified(true)
		needsUpdate = true
	}

	googleProviderFound := false
	for _, provider := range existing.ProviderUserInfo {
		if provider == nil || provider.ProviderID != "google.com" {
			continue
		}
		googleProviderFound = true
		if provider.UID != common.DevelopmentUserID {
			return fmt.Errorf("development Firebase user %q has conflicting Google provider UID %q", common.DevelopmentUserID, provider.UID)
		}
	}
	if !googleProviderFound {
		update.ProviderToLink(&auth.UserProvider{
			UID:         common.DevelopmentUserID,
			ProviderID:  "google.com",
			Email:       common.DevelopmentUserEmail,
			DisplayName: common.DevelopmentUserName,
		})
		needsUpdate = true
	}

	if !needsUpdate {
		log.Printf("event=development_auth_user action=already_exists uid=%q", common.DevelopmentUserID)
		return nil
	}
	if _, err := client.UpdateUser(ctx, common.DevelopmentUserID, update); err != nil {
		return fmt.Errorf("reconcile development Firebase user: %w", err)
	}
	log.Printf("event=development_auth_user action=updated uid=%q", common.DevelopmentUserID)
	return nil
}
