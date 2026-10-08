package main

import (
	"net"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// ServerHeader adds a custom Server response header.
func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		c.Response().Header().Set("Server", "ACS-4210-Echo")
		return next(c)
	}
}

// InsideTheBuilding labels the request as internal (loopback/private) or not.
// It only labels; it never blocks the request.
func InsideTheBuilding(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		inside := false
		via := "unparseable"

		if ip := net.ParseIP(c.RealIP()); ip != nil {
			inside = ip.IsLoopback() || ip.IsPrivate()
			via = "realip+private/loopback"
		}

		c.Set("inside", inside)
		c.Set("via", via)

		if inside {
			c.Response().Header().Set("X-Inside-The-Building", "1")
		} else {
			c.Response().Header().Set("X-Inside-The-Building", "0")
		}

		return next(c)
	}
}

func main() {
	e := echo.New()

	// Use the direct network peer; do not trust X-Forwarded-For or X-Real-IP.
	e.IPExtractor = echo.ExtractIPDirect()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(InsideTheBuilding)
	e.Use(ServerHeader)

	e.GET("/whoami", func(c *echo.Context) error {
		inside, _ := c.Get("inside").(bool)
		via, _ := c.Get("via").(string)
		return c.JSON(http.StatusOK, map[string]any{
			"ip":     c.RealIP(),
			"inside": inside,
			"via":    via,
		})
	})

	e.Logger.Error("server stopped", "error", e.Start(":1323"))
}
