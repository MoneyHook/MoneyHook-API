package router

import (
	common "MoneyHook/MoneyHook-API/common"
	"context"
	"errors"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

func NewFirebaseAuth() (*auth.Client, error) {
	ctx := context.Background()
	projectID := common.GetEnv("GOOGLE_CLOUD_PROJECT", common.GetEnv("GCLOUD_PROJECT", ""))
	emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if emulatorHost != "" && projectID == "" {
		return nil, errors.New("GOOGLE_CLOUD_PROJECT is required when using the Firebase Auth Emulator")
	}

	var config *firebase.Config
	if projectID != "" {
		config = &firebase.Config{ProjectID: projectID}
	}
	var opts []option.ClientOption
	credentialsPath := common.GetEnv("SECRET_PATH", common.GetEnv("GOOGLE_APPLICATION_CREDENTIALS", ""))
	if credentialsPath != "" {
		if _, err := os.Stat(credentialsPath); err != nil {
			return nil, fmt.Errorf("read Firebase credentials %q: %w", credentialsPath, err)
		}
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
	}

	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase app: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase Auth client: %w", err)
	}
	return client, nil
}
