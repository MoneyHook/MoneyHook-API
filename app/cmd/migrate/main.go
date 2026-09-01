package main

import (
	"MoneyHook/MoneyHook-API/db"
	"context"
	"log"
)

func main() {
	if err := db.Migrate(context.Background()); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
}
