package api

import (
	"fmt"

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

	dataset.POST("/:dataset/content", func(c echo.Context) error {
		var content []core.Content
		var dataset core.Dataset
		results := struct {
			success []string
			fail    []string
		}{
			success: make([]string, 0),
			fail:    make([]string, 0),
		}

		d := c.Param("dataset")

		if d == "" {
			return fmt.Errorf("dataset must be specified")
		}

		res := db.Where("name = ?", d).First(&dataset)
		if res.Error != nil {
			return res.Error
		}

		if err := c.Bind(&content); err != nil {
			return err
		}

		for _, cnt := range content {
			res := db.Create(&cnt)
			if res.Error != nil {
				results.fail = append(results.fail, cnt.CommP)
				continue
			}

			dataset.Contents = append(dataset.Contents, cnt)
			results.success = append(results.success, cnt.CommP)
		}

		res = db.Save(&dataset)

		if res.Error != nil {
			return res.Error
		}

		return c.JSON(200, results)
	})

}
