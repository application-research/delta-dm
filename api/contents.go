package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/application-research/delta-dm/core"
	db "github.com/application-research/delta-dm/db"
	"github.com/jszwec/csvutil"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ConfigureContentsRouter(e *echo.Group, dldm *core.DeltaDM) {
	contents := e.Group("/contents")

	contents.Use(dldm.AS.AuthMiddleware)

	contents.GET("/:dataset", func(c echo.Context) error {
		var content []db.Content
		var dataset db.Dataset

		d := c.Param("dataset")

		if d == "" {
			return fmt.Errorf("dataset id must be specified")
		}
		did, err := strconv.ParseUint(d, 10, 64)
		if err != nil {
			return fmt.Errorf("dataset id must be numeric %s", err)
		}

		if tx := dldm.DB.First(&dataset, did); tx.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("dataset not found")
			}
			return fmt.Errorf("failed to get dataset: %s", tx.Error)
		}

		err = dldm.DB.Model(&dataset).Association("Contents").Find(&content)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, content)
	})

	contents.POST("/:dataset", func(c echo.Context) error {
		var content []db.Content
		var dataset db.Dataset
		results := struct {
			Success []string `json:"success"`
			Fail    []string `json:"fail"`
		}{
			Success: make([]string, 0),
			Fail:    make([]string, 0),
		}

		d := c.Param("dataset")

		if d == "" {
			return fmt.Errorf("dataset id must be specified")
		}
		did, err := strconv.ParseUint(d, 10, 64)
		if err != nil {
			return fmt.Errorf("dataset id must be numeric %s", err)
		}

		// Check if dataset exists
		if tx := dldm.DB.First(&dataset, did); tx.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("dataset not found")
			}
			return fmt.Errorf("failed to get dataset: %s", tx.Error)
		}

		it := c.QueryParam("import_type")
		if it == "singularity" {
			var sContent []SingularityJSON
			if err := c.Bind(&sContent); err != nil {
				return err
			}

			// Marshal into core.Content
			for _, cnt := range sContent {
				content = append(content, cnt.toDeltaContent())
			}
		} else if it == "csv" {
			csvBytes, err := ioutil.ReadAll(c.Request().Body)
			if err != nil {
				return err
			}

			if err := csvutil.Unmarshal(csvBytes, &content); err != nil {
				return fmt.Errorf("error parsing csv : %s", err)
			}

		} else {
			if err := c.Bind(&content); err != nil {
				return err
			}
		}

		for _, cnt := range content {
			// Check for bad data
			if cnt.CommP == "" || cnt.PayloadCID == "" || cnt.PaddedSize == 0 || cnt.Size == 0 {
				results.Fail = append(results.Fail, cnt.CommP)
				continue
			}

			cnt.DatasetID = dataset.ID

			err := dldm.DB.Create(&cnt).Error
			if err != nil {
				results.Fail = append(results.Fail, cnt.CommP)
				continue
			}

			results.Success = append(results.Success, cnt.CommP)
		}

		return c.JSON(http.StatusOK, results)
	})
}

// Field name mapping for JSON exported from singularity db
type SingularityJSON struct {
	CarSize   uint64 `json:"carSize"`
	DataCid   string `json:"dataCid"`
	PieceCid  string `json:"pieceCid"`
	PieceSize uint64 `json:"pieceSize"`
}

func (s *SingularityJSON) toDeltaContent() db.Content {
	return db.Content{
		CommP:      s.PieceCid,
		PayloadCID: s.DataCid,
		Size:       s.CarSize,
		PaddedSize: s.PieceSize,
	}
}
