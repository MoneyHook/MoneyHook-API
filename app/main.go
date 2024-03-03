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
	h := handler.NewHandler(cs)
	h.Register(v1)

	e.Logger.Fatal(e.Start(":8080"))
}
