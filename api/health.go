package api

import (
	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
)

// Health check routes for verifying API is alive
func ConfigureHealthRouter(e *echo.Group, dldm *core.DeltaDM) {
	health := e.Group("/health")

	health.GET("", func(c echo.Context) error {
		err := RequestAuthHeaderCheck(c)
		if err != nil {
			return c.JSON(401, err.Error())
		}

		return c.JSON(200, "alive")
	})
}
