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
	var d PostWalletBody

	if err := c.Bind(&d); err != nil {
		return err
	}

	deltaResp, err := dldm.DAPI.AddWallet(core.AddWalletRequest(d))
	if err != nil {
		return fmt.Errorf("could not add wallet %s", err)
	}

	newWallet := core.Wallet{
		Address: deltaResp.WalletAddr,
		Type:    d.Type,
	}

	res := dldm.DB.Model(core.Wallet{}).Create(&newWallet)
	if res.Error != nil {
		return res.Error
	}

	return c.JSON(200, newWallet)
}
