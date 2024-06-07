package router

import (
	common "MoneyHook/MoneyHook-API/common"
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

func NewFirebaseAuth() *auth.Client {
	log.Printf("Start Firebase Setup")
	creds := common.GetEnv("SECRET_PATH", "")
	if creds == "" {
		log.Fatalf("'SECRET_PATH' is not found.")
	}
	opt := option.WithCredentialsFile(creds)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
	log.Printf("Finish Firebase Setup")

	return client
}
