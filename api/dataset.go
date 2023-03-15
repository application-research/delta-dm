package api

import (
	"fmt"
	"io/ioutil"

	"github.com/application-research/delta-dm/core"
	"github.com/jszwec/csvutil"
	"github.com/labstack/echo/v4"
)

func ConfigureDatasetRouter(e *echo.Group, dldm *core.DeltaDM) {
	dataset := e.Group("/datasets")

	dataset.GET("", func(c echo.Context) error {
		var ds []core.Dataset

		dldm.DB.Model(&core.Dataset{}).Preload("Wallet").Find(&ds)

		// Find  # of bytes total and replicated for each dataset
		for i, d := range ds {
			var rb [2]uint64
			dldm.DB.Raw("select SUM(size) s, SUM(padded_size) ps FROM contents c inner join replications r on r.content_comm_p = c.comm_p where r.status = 'SUCCESS' AND dataset_name = ?", d.Name).Row().Scan(&rb[0], &rb[1])

			var tb [2]uint64
			dldm.DB.Raw("select SUM(size) s, SUM(padded_size) ps FROM contents where dataset_name = ?", d.Name).Row().Scan(&tb[0], &tb[1])

			ds[i].BytesReplicated = core.ByteSizes{Raw: rb[0], Padded: rb[1]}
			ds[i].BytesTotal = core.ByteSizes{Raw: tb[0], Padded: tb[1]}
		}

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

		res := dldm.DB.Create(&ads)

		if res.Error != nil {
			return res.Error
		}

		return c.JSON(200, ads)
	})

	dataset.GET("/content/:dataset", func(c echo.Context) error {
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

	dataset.POST("/content/:dataset", func(c echo.Context) error {
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

		// Check if dataset exists
		res := dldm.DB.Where("name = ?", d).First(&dataset)
		if res.Error != nil {
			return fmt.Errorf("could not find dataset %s : %s", d, res.Error)
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

			cnt.DatasetName = dataset.Name

			err := dldm.DB.Create(&cnt).Error
			if err != nil {
				results.Fail = append(results.Fail, cnt.CommP)
				continue
			}

			results.Success = append(results.Success, cnt.CommP)
		}

		return c.JSON(200, results)
	})
}

// Field name mapping for JSON exported from singularity db
type SingularityJSON struct {
	CarSize   uint64 `json:"carSize"`
	DataCid   string `json:"dataCid"`
	PieceCid  string `json:"pieceCid"`
	PieceSize uint64 `json:"pieceSize"`
}

func (s *SingularityJSON) toDeltaContent() core.Content {
	return core.Content{
		CommP:      s.PieceCid,
		PayloadCID: s.DataCid,
		Size:       s.CarSize,
		PaddedSize: s.PieceSize,
	}
}
