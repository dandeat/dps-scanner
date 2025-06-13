package routes

import (
	"dps-scanner-gateout/services"
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func RoutesApi(e *echo.Echo, usecaseSvc services.UsecaseService) {
	// Middleware
	routePrefix := e.Group("/")

	routePrefix.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		log.Println("[Start]")
		log.Println("EndPoint :", c.Path())
		log.Println("Header :", c.Request().Header)
		log.Println("Body :", string(reqBody))
		log.Println("Response :", string(resBody))
		log.Println("[End]")
	}))

	// WS
	// routePrefix.GET("/ws", usecaseSvc.WSHandler)
}
