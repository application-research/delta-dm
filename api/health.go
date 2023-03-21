package api

import (
	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
)

// Health check routes for verifying API is alive
func ConfigureHealthRouter(e *echo.Group, dldm *core.DeltaDM) {
	health := e.Group("/health")

	health.Use(dldm.AS.AuthMiddleware)

	health.GET("", func(c echo.Context) error {
		uuid := dldm.DAPI.NodeUUID

		return c.JSON(200, uuid)
	})
}
