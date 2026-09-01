package router

import (
	common "MoneyHook/MoneyHook-API/common"
	"context"
	"errors"
	"strings"
	"testing"

	"firebase.google.com/go/v4/auth"
)

type fakeAuthAdminClient struct {
	user          *auth.UserRecord
	getErr        error
	importedUsers []*auth.UserToImport
	importResult  *auth.UserImportResult
	importErr     error
	updatedUID    string
	update        *auth.UserToUpdate
	updateErr     error
}

func (f *fakeAuthAdminClient) GetUser(context.Context, string) (*auth.UserRecord, error) {
	return f.user, f.getErr
}

func (f *fakeAuthAdminClient) ImportUsers(_ context.Context, users []*auth.UserToImport, _ ...auth.UserImportOption) (*auth.UserImportResult, error) {
	f.importedUsers = users
	return f.importResult, f.importErr
}

func (f *fakeAuthAdminClient) UpdateUser(_ context.Context, uid string, update *auth.UserToUpdate) (*auth.UserRecord, error) {
	f.updatedUID = uid
	f.update = update
	return nil, f.updateErr
}

func TestDevelopmentUserEnabledFromEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		exists  bool
		want    bool
		wantErr bool
	}{
		{name: "unset", exists: false},
		{name: "empty", value: "", exists: true},
		{name: "true", value: " true ", exists: true, want: true},
		{name: "false", value: "false", exists: true},
		{name: "invalid", value: "enabled", exists: true, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseDevelopmentUserEnabled(test.value, test.exists)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseDevelopmentUserEnabled() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("parseDevelopmentUserEnabled() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestEnsureDevelopmentUserCreatesMissingUser(t *testing.T) {
	client := &fakeAuthAdminClient{
		getErr:       errors.New("user not found"),
		importResult: &auth.UserImportResult{SuccessCount: 1},
	}

	err := ensureDevelopmentUser(context.Background(), client, func(error) bool { return true })
	if err != nil {
		t.Fatalf("ensureDevelopmentUser() error = %v", err)
	}
	if len(client.importedUsers) != 1 {
		t.Fatalf("ImportUsers() received %d users, want 1", len(client.importedUsers))
	}
	if client.updatedUID != "" {
		t.Fatal("UpdateUser() should not be called when creating a missing user")
	}
}

func TestEnsureDevelopmentUserIsIdempotent(t *testing.T) {
	client := &fakeAuthAdminClient{user: developmentUserRecord()}

	err := ensureDevelopmentUser(context.Background(), client, func(error) bool { return false })
	if err != nil {
		t.Fatalf("ensureDevelopmentUser() error = %v", err)
	}
	if len(client.importedUsers) != 0 || client.updatedUID != "" {
		t.Fatal("an already-correct user should not be imported or updated")
	}
}

func TestEnsureDevelopmentUserCorrectsExistingUser(t *testing.T) {
	client := &fakeAuthAdminClient{user: &auth.UserRecord{
		UserInfo:      &auth.UserInfo{UID: common.DevelopmentUserID, DisplayName: "旧表示名", Email: "old@example.com"},
		EmailVerified: false,
	}}

	err := ensureDevelopmentUser(context.Background(), client, func(error) bool { return false })
	if err != nil {
		t.Fatalf("ensureDevelopmentUser() error = %v", err)
	}
	if client.updatedUID != common.DevelopmentUserID || client.update == nil {
		t.Fatal("UpdateUser() should correct an existing user")
	}
}

func TestEnsureDevelopmentUserRejectsProviderUIDConflict(t *testing.T) {
	client := &fakeAuthAdminClient{user: &auth.UserRecord{
		UserInfo:         &auth.UserInfo{UID: common.DevelopmentUserID},
		ProviderUserInfo: []*auth.UserInfo{{UID: "different-google-uid", ProviderID: "google.com"}},
	}}

	err := ensureDevelopmentUser(context.Background(), client, func(error) bool { return false })
	if err == nil || !strings.Contains(err.Error(), "conflicting Google provider UID") {
		t.Fatalf("ensureDevelopmentUser() error = %v, want provider UID conflict", err)
	}
	if client.updatedUID != "" {
		t.Fatal("UpdateUser() should not be called for a provider UID conflict")
	}
}

func TestEnsureDevelopmentUserFailsOnUnexpectedLookupError(t *testing.T) {
	client := &fakeAuthAdminClient{getErr: errors.New("auth unavailable")}

	err := ensureDevelopmentUser(context.Background(), client, func(error) bool { return false })
	if err == nil || !strings.Contains(err.Error(), "lookup development Firebase user") {
		t.Fatalf("ensureDevelopmentUser() error = %v, want lookup failure", err)
	}
}

func TestEnsureDevelopmentUserFailsOnImportError(t *testing.T) {
	client := &fakeAuthAdminClient{getErr: errors.New("user not found"), importErr: errors.New("auth unavailable")}

	err := ensureDevelopmentUser(context.Background(), client, func(error) bool { return true })
	if err == nil || !strings.Contains(err.Error(), "create development Firebase user") {
		t.Fatalf("ensureDevelopmentUser() error = %v, want import failure", err)
	}
}

func developmentUserRecord() *auth.UserRecord {
	return &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:         common.DevelopmentUserID,
			DisplayName: common.DevelopmentUserName,
			Email:       common.DevelopmentUserEmail,
		},
		EmailVerified: true,
		ProviderUserInfo: []*auth.UserInfo{{
			UID:        common.DevelopmentUserID,
			ProviderID: "google.com",
			Email:      common.DevelopmentUserEmail,
		}},
	}
}
