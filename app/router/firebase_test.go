package router

import (
	"strings"
	"testing"
)

func TestNewFirebaseAuthRequiresProjectForEmulator(t *testing.T) {
	t.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "127.0.0.1:9099")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GCLOUD_PROJECT", "")
	t.Setenv("SECRET_PATH", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

	_, err := NewFirebaseAuth()
	if err == nil || !strings.Contains(err.Error(), "GOOGLE_CLOUD_PROJECT") {
		t.Fatalf("error=%v, want missing project error", err)
	}
}

func TestNewFirebaseAuthRejectsMissingCredentialsFile(t *testing.T) {
	t.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "demo-moneyhooks")
	t.Setenv("SECRET_PATH", "/definitely/missing/firebase-credentials.json")

	_, err := NewFirebaseAuth()
	if err == nil || !strings.Contains(err.Error(), "read Firebase credentials") {
		t.Fatalf("error=%v, want credentials error", err)
	}
}
