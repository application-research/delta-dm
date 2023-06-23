package api

import (
	"fmt"
	"net/http"

	"github.com/application-research/delta-dm/core"
	db "github.com/application-research/delta-dm/db"
	"github.com/application-research/delta-dm/util"
	"github.com/labstack/echo/v4"
)

type DatasetPutBody struct {
	Name string `json:"name"`
}

func ConfigureDatasetsRouter(e *echo.Group, dldm *core.DeltaDM) {
	datasets := e.Group("/datasets")

	datasets.Use(dldm.AS.AuthMiddleware)

	datasets.GET("", func(c echo.Context) error {
		var ds []db.Dataset

		dldm.DB.Preload("Wallets").Preload("ReplicationProfiles").Find(&ds)

		// Find  # of bytes total and replicated for each dataset
		for i, d := range ds {
			var rb [2]uint64
			dldm.DB.Raw("select SUM(size) s, SUM(padded_size) ps FROM contents c inner join replications r on r.content_comm_p = c.comm_p where r.status = 'SUCCESS' AND dataset_id = ?", d.ID).Row().Scan(&rb[0], &rb[1])

			var tb [2]uint64
			dldm.DB.Raw("select SUM(size) s, SUM(padded_size) ps FROM contents where dataset_id = ?", d.ID).Row().Scan(&tb[0], &tb[1])

			ds[i].BytesReplicated = db.ByteSizes{Raw: rb[0], Padded: rb[1]}
			ds[i].BytesTotal = db.ByteSizes{Raw: tb[0], Padded: tb[1]}

			var countReplicated uint64 = 0
			var countTotal uint64 = 0
			dldm.DB.Raw("select count(*) cr FROM contents c inner join replications r on r.content_comm_p = c.comm_p where r.status = 'SUCCESS' AND dataset_id = ?", d.ID).Row().Scan(&countReplicated)
			dldm.DB.Raw("select count(*) cr FROM contents c where dataset_id = ?", d.ID).Row().Scan(&countTotal)

			ds[i].CountReplicated = countReplicated
			ds[i].CountTotal = countTotal
		}

		return c.JSON(http.StatusOK, ds)
	})

	datasets.POST("", func(c echo.Context) error {
		var ads db.Dataset

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

		if !util.ValidateDatasetName(ads.Name) {
			return fmt.Errorf("invalid dataset name. must contain only lowercase letters, numbers and hyphens. must begin and end with a letter. must not contain consecutive hyphens")
		}

		res := dldm.DB.Create(&ads)

		if res.Error != nil {
			if res.Error.Error() == "UNIQUE constraint failed: datasets.name" {
				return fmt.Errorf("dataset with name %s already exists", ads.Name)
			}
			return res.Error
		}

		return c.JSON(http.StatusOK, ads)
	})

	datasets.PUT("/:dataset_id", func(c echo.Context) error {
		did := c.Param("dataset_id")
		if did == "" {
			return fmt.Errorf("dataset id not specified")
		}

		var d DatasetPutBody

		if err := c.Bind(&d); err != nil {
			return err
		}

		var existing db.Dataset
		res := dldm.DB.Model(&db.Dataset{}).Where("id = ?", did).First(&existing)

		if res.Error != nil {
			return fmt.Errorf("error fetching dataset %s", res.Error)
		}

		if d.Name != "" {
			existing.Name = d.Name
		}

		res = dldm.DB.Save(&existing)
		if res.Error != nil {
			return fmt.Errorf("error saving dataset %s", res.Error)
		}

		return c.JSON(http.StatusOK, existing)
	})
}
