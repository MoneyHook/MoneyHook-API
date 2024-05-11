package main

import (
	"MoneyHook/MoneyHook-API/db"
	"MoneyHook/MoneyHook-API/handler"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/store"
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"http://localhost:3000"},
	// 	AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	// }))
	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Success, running")
	})

	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	v1 := e.Group("/api")

	d := db.New()

	us := store.NewUserStore(d)
	ts := store.NewTransactionStore(d)
	fs := store.NewFixedStore(d)
	cs := store.NewCategoryStore(d)
	scs := store.NewSubCategoryStore(d)
	h := handler.NewHandler(client, us, ts, fs, cs, scs)
	h.Register(v1)

	message.Read()

	e.Logger.Fatal(e.Start(":8080"))
}
