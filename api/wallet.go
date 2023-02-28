package api

import (
	"fmt"

	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
)

func ConfigureWalletRouter(e *echo.Group, dldm *core.DeltaDM) {
	replication := e.Group("/wallet")

	replication.GET("", func(c echo.Context) error {

		p := c.QueryParam("provider")
		ds := c.QueryParam("dataset")

		var r []core.Replication

		tx := dldm.DB.Model(&core.Replication{}).Joins("inner join contents c on c.comm_p = replications.content_comm_p").Joins("inner join datasets d on d.id = c.dataset_id")

		if ds != "" {
			tx.Where("d.name = ?", ds)
		}

		if p != "" {
			tx.Where("replications.provider_actor_id = ?", p)
		}

		tx.Find(&r)

		return c.JSON(200, r)
	})

	replication.POST("", func(c echo.Context) error {
		return handlePostWallet(c, dldm)
	})
}

// We will not private key in DDM DB, pass it on to Delta
type PostWalletBody struct {
	Type       string `json:"Type"`
	PrivateKey string `json:"PrivateKey"`
}

// POST /api/wallet
// @description add/import a wallet
// @returns newly added wallet info
func handlePostWallet(c echo.Context, dldm *core.DeltaDM) error {
	var w PostWalletBody

	if err := c.Bind(&w); err != nil {
		return err
	}

	ds := c.QueryParam("dataset")

	var exists bool
	err := dldm.DB.Model(core.Dataset{}).
		Select("count(*) > 0").
		Where("name = ?", ds).
		Find(&exists).
		Error

	if err != nil {
		return fmt.Errorf("could not check if dataset %s exists: %s", ds, err)
	}

	if !exists {
		return fmt.Errorf("dataset %s does not exist", ds)
	}

	deltaResp, err := dldm.DAPI.AddWallet(core.AddWalletRequest(w))
	if err != nil {
		return fmt.Errorf("could not add wallet %s", err)
	}

	newWallet := core.Wallet{
		Addr:        deltaResp.WalletAddr,
		Type:        w.Type,
		DatasetName: ds,
	}

	// Create a new record, or update existing record for the dataset if one already exists
	res := dldm.DB.Model(core.Wallet{}).Where("dataset_name = ?", ds).Assign(newWallet).FirstOrCreate(&newWallet)
	if res.Error != nil {
		return res.Error
	}

	return c.JSON(200, newWallet)
}
