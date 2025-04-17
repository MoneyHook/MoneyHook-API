package main

import (
	common "MoneyHook/MoneyHook-API/common"
	"MoneyHook/MoneyHook-API/db"
	"MoneyHook/MoneyHook-API/handler"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/router"
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

	// e.Use(middleware.KeyAuthWithConfig(router.ValidateJwt()))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Success, running")
	})

	// e.GET("/generateKey", func(c echo.Context) error {
	// return router.GenerateJWT(c)
	// })

	v1 := e.Group("/api")

	d := db.New()
	client := router.NewFirebaseAuth()
	h := handler.NewHandler(client, d.UserStore, d.TransactionStore, d.FixedStore, d.CategoryStore, d.SubCategoryStore, d.PaymentResourceStore, d.JobsStore)
	h.Register(v1)

	message.Read()

	e.Logger.Fatal(e.Start(":8080"))
}
