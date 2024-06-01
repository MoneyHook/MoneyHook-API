package main

import (
	"MoneyHook/MoneyHook-API/db"
	"MoneyHook/MoneyHook-API/handler"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/router"
	"MoneyHook/MoneyHook-API/store"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	log.Printf("Start Application")
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders: []string{"Content-Length"},
	}))

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

	client := router.NewFirebaseAuth()
	h := handler.NewHandler(client, us, ts, fs, cs, scs)
	h.Register(v1)

	message.Read()

	e.Logger.Fatal(e.Start(":8080"))
}
