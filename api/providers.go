package api

import (
	"fmt"
	"net/http"

	"github.com/application-research/delta-dm/core"
	"github.com/filecoin-project/go-address"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ProviderPutBody struct {
	ActorName        string   `json:"actor_name"`
	AllowSelfService string   `json:"allow_self_service"`
	AllowedDatasets  []string `json:"allowed_datasets"`
}

func ConfigureProvidersRouter(e *echo.Group, dldm *core.DeltaDM) {
	providers := e.Group("/providers")

	providers.Use(dldm.AS.AuthMiddleware)

	providers.GET("", func(c echo.Context) error {
		var p []core.Provider

		dldm.DB.Preload("AllowedDatasets").Find(&p)

		for i, sp := range p {
			var rb [2]uint64
			dldm.DB.Raw("select SUM(size) s, SUM(padded_size) ps FROM contents c inner join replications r on r.content_comm_p = c.comm_p where r.status = 'SUCCESS' AND r.provider_actor_id = ?", sp.ActorID).Row().Scan(&rb[0], &rb[1])

			p[i].BytesReplicated = core.ByteSizes{Raw: rb[0], Padded: rb[1]}
		}

		return c.JSON(200, p)
	})

	providers.POST("", func(c echo.Context) error {
		var p core.Provider

		if err := c.Bind(&p); err != nil {
			return err
		}

		// Check to ensure the actor id is valid
		_, err := address.NewFromString(p.ActorID)
		if err != nil {
			return fmt.Errorf("invalid actor id %s: %s", p.ActorID, err)
		}

		p.Key = uuid.New()

		res := dldm.DB.Create(&p)

		if res.Error != nil {
			return res.Error
		}
		return c.JSON(http.StatusOK, p)
	})

	providers.PUT("/:provider_id", func(c echo.Context) error {
		pid := c.Param("provider_id")
		if pid == "" {
			return fmt.Errorf("provider id not specified")
		}

		var p ProviderPutBody

		if err := c.Bind(&p); err != nil {
			return err
		}

		var existing core.Provider
		res := dldm.DB.Model(&core.Provider{}).Where("actor_id = ?", pid).First(&existing)

		if res.Error != nil {
			return fmt.Errorf("error fetching provider %s", res.Error)
		}

		if p.ActorName != "" {
			existing.ActorName = p.ActorName
		}

		if p.AllowSelfService == "on" {
			existing.AllowSelfService = true
		} else if p.AllowSelfService == "off" {
			existing.AllowSelfService = false
		}

		// If array of allowed datasets is empty, don't modify the association
		if len(p.AllowedDatasets) > 0 {
			var newAllowedDatasets []core.Dataset
			for _, ds_name := range p.AllowedDatasets {
				var ds core.Dataset
				res := dldm.DB.Model(&core.Dataset{}).Where("name = ?", ds_name).First(&ds)
				if res.Error != nil {
					return fmt.Errorf("error fetching dataset %s : %s", ds_name, res.Error)
				} else {
					newAllowedDatasets = append(newAllowedDatasets, ds)
				}
			}

			dldm.DB.Model(&existing).Association("AllowedDatasets").Replace(newAllowedDatasets)
		}

		res = dldm.DB.Save(&existing)
		if res.Error != nil {
			return fmt.Errorf("error saving provider %s", res.Error)
		}

		return c.JSON(http.StatusOK, existing)
	})

}
