package api

import (
	"fmt"

	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
)

func ConfigureDatasetRouter(e *echo.Group, dldm *core.DeltaDM) {
	dataset := e.Group("/dataset")

	dataset.GET("", func(c echo.Context) error {
		var ds []core.Dataset

		dldm.DB.Model(&core.Dataset{}).Preload("Wallet").Find(&ds)

		return c.JSON(200, ds)
	})

	dataset.POST("", func(c echo.Context) error {
		var ads core.Dataset

		if err := c.Bind(&ads); err != nil {
			return err
		}

		if ads.ReplicationQuota < 1 {
			return fmt.Errorf("replication quota must be greater than 0")
		}

		// Bound deal durations between 180 and 540

		if ads.DealDuration < 180 || ads.DealDuration > 540 {
			return fmt.Errorf("deal duration must be between 180 and 540 days")
		}

		if ads.DelayStartEpoch < 1 || ads.DealDuration > 14 {
			return fmt.Errorf("delay start epoch must be between 1 and 14 days")
		}

		res := dldm.DB.Create(&ads)

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

		dldm.DB.Where("name = ?", d).First(&dataset)

		err := dldm.DB.Model(&dataset).Association("Contents").Find(&content)
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

		res := dldm.DB.Where("name = ?", d).First(&dataset)
		if res.Error != nil {
			return res.Error
		}

		if err := c.Bind(&content); err != nil {
			return err
		}

		for _, cnt := range content {
			err := dldm.DB.Create(&cnt).Error
			if err != nil {
				results.Fail = append(results.Fail, cnt.CommP)
				continue
			}

			results.Success = append(results.Success, cnt.CommP)
			err = dldm.DB.Model(&dataset).Association("Contents").Append(&cnt)
			if err != nil {
				return err
			}
		}

		return c.JSON(200, results)
	})
}
