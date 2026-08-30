package main

import (
	common "MoneyHook/MoneyHook-API/common"
	"MoneyHook/MoneyHook-API/db"
	"MoneyHook/MoneyHook-API/handler"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/router"
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	log.Printf("Start Application")
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{common.GetEnv("FRONT_URL", "http://localhost:3000")},
		AllowMethods:  []string{echo.GET, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders: []string{"Content-Length"},
	}))

	registerHealthRoute(e)

	api := e.Group("/api")

	seedDataEnabled, err := db.SeedDataEnabledFromEnvironment()
	if err != nil {
		log.Fatalf("Database setup failed: %v", err)
	}
	developmentUserEnabled, err := router.DevelopmentUserEnabledFromEnvironment()
	if err != nil {
		log.Fatalf("Development user setup failed: %v", err)
	}
	client, err := router.NewFirebaseAuth()
	if err != nil {
		log.Fatalf("Firebase setup failed: %v", err)
	}
	if seedDataEnabled && developmentUserEnabled {
		if err := router.EnsureDevelopmentUser(context.Background(), client); err != nil {
			log.Fatalf("Development user setup failed: %v", err)
		}
	}
	d, err := db.New()
	if err != nil {
		log.Fatalf("Database setup failed: %v", err)
	}
	h := handler.New(handler.Dependencies{
		FirebaseClient:       client,
		UserStore:            d.UserStore,
		TransactionStore:     d.TransactionStore,
		FixedStore:           d.FixedStore,
		CategoryStore:        d.CategoryStore,
		SubCategoryStore:     d.SubCategoryStore,
		PaymentResourceStore: d.PaymentResourceStore,
		JobStore:             d.JobStore,
	})
	h.Register(api)

	message.Read()

	e.Logger.Fatal(e.Start(":8080"))
}

func registerHealthRoute(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Success, running")
	})
}
