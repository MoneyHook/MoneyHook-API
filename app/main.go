package main

import (
	"net/http"

	"MoneyHook/MoneyHook-API/db"
	"MoneyHook/MoneyHook-API/handler"
	"MoneyHook/MoneyHook-API/message"
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
