package router

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

func NewFirebaseAuth() *auth.Client {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	return client
}
