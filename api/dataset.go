package api

import (
	"github.com/application-research/delta-ldm/core"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ConfigureDatasetRouter(e *echo.Group, db *gorm.DB) {
	dataset := e.Group("/dataset")

	dataset.GET("", func(c echo.Context) error {
		var ds []core.Dataset

		db.Find(&ds)

		return c.JSON(200, ds)
	})

	dataset.POST("", func(c echo.Context) error {
		var ads core.Dataset

		if err := c.Bind(&ads); err != nil {
			return err
		}

		res := db.Create(&ads)

		if res.Error != nil {
			return res.Error
		}

		return c.JSON(200, ads)
	})

}
