package api

import (
	"net/http"

	"github.com/application-research/delta-ldm/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ConfigureProvidersRouter(e *echo.Group, db *gorm.DB) {
	providers := e.Group("/providers")

	providers.GET("", func(c echo.Context) error {
		var p []core.Provider

		db.Find(&p)

		return c.JSON(200, p)
	})

	providers.POST("", func(c echo.Context) error {
		var p core.Provider

		if err := c.Bind(&p); err != nil {
			return err
		}

		p.Key = uuid.New()

		res := db.Create(&p)

		if res.Error != nil {
			return res.Error
		}
		return c.JSON(http.StatusOK, p)
	})

}
