package router

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

func NewFirebaseAuth() *auth.Client {
	log.Printf("Start Firebase Setup")
	app, err := firebase.NewApp(context.Background(), nil)
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
