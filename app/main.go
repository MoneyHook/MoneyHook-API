package main

import (
	"net/http"

	"MoneyHook/MoneyHook-API/db"
	"MoneyHook/MoneyHook-API/handler"
	"MoneyHook/MoneyHook-API/store"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Success, running")
	})

	v1 := e.Group("/api")

	d := db.New()

	cs := store.NewCategoryStore(d)
	scs := store.NewSubCategoryStore(d)
	ts := store.NewTransactionStore(d)
	fs := store.NewFixedStore(d)
	h := handler.NewHandler(cs, scs, ts, fs)
	h.Register(v1)

	e.Logger.Fatal(e.Start(":8080"))
}
