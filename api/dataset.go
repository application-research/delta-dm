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

	dataset.GET("/:dataset/content", func(c echo.Context) error {
		var content []core.Content
		var dataset core.Dataset
		d := c.Param("dataset")

		if d == "" {
			return fmt.Errorf("dataset must be specified")
		}

		db.Where("name = ?", d).First(&dataset)

		err := db.Model(&dataset).Association("Contents").Find(&content)
		if err != nil {
			return err
		}

		return c.JSON(200, content)
	})

	dataset.POST("/:dataset/content", func(c echo.Context) error {
		var content []core.Content
		var dataset core.Dataset
		results := struct {
			Success []string `json:"success"`
			Fail    []string `json:"fail"`
		}{
			Success: make([]string, 0),
			Fail:    make([]string, 0),
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
			err := db.Create(&cnt).Error
			if err != nil {
				results.Fail = append(results.Fail, cnt.CommP)
				continue
			}

			results.Success = append(results.Success, cnt.CommP)
			err = db.Model(&dataset).Association("Contents").Append(&cnt)
			if err != nil {
				return err
			}
		}

		return c.JSON(200, results)
	})
}
