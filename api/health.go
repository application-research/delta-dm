package api

import (
	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
	"net/http"
)

// Health check routes for verifying API is alive
func ConfigureHealthRouter(e *echo.Group, dldm *core.DeltaDM) {
	health := e.Group("/health")

	health.Use(dldm.AS.AuthMiddleware)

	health.GET("", func(c echo.Context) error {

		resp := struct {
			UUID      string              `json:"uuid"`
			DDMInfo   core.DeploymentInfo `json:"ddm_info"`
			DeltaInfo core.DeploymentInfo `json:"delta_info"`
		}{
			UUID:      dldm.DAPI.NodeUUID,
			DDMInfo:   dldm.Info,
			DeltaInfo: dldm.DAPI.DeltaDeploymentInfo,
		}

		return c.JSON(http.StatusOK, resp)
	})
}
