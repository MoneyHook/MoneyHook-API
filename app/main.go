package main

import (
	"MoneyHook/MoneyHook-API/db"
	"MoneyHook/MoneyHook-API/handler"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/store"
	"net/http"

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

	v1 := e.Group("/api")

	d := db.New()

	us := store.NewUserStore(d)
	ts := store.NewTransactionStore(d)
	fs := store.NewFixedStore(d)
	cs := store.NewCategoryStore(d)
	scs := store.NewSubCategoryStore(d)
	h := handler.NewHandler(us, ts, fs, cs, scs)
	h.Register(v1)

	message.Read()

	e.Logger.Fatal(e.Start(":8080"))
}
