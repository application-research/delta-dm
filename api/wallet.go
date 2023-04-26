package api

import (
	"fmt"

	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
)

func ConfigureWalletsRouter(e *echo.Group, dldm *core.DeltaDM) {
	wallets := e.Group("/wallets")

	wallets.Use(dldm.AS.AuthMiddleware)

	wallets.GET("", func(c echo.Context) error {
		authKey := c.Get(core.AUTH_KEY).(string)

		ds := c.QueryParam("dataset")

		var w []core.Wallet

		tx := dldm.DB.Model(&core.Wallet{}).Preload("Datasets")

		if ds != "" {
			tx.Where("dataset_name = ?", ds)
		}

		tx.Find(&w)

		for i, wallet := range w {
			bal, err := dldm.DAPI.GetWalletBalance(wallet.Addr, authKey)
			if err != nil {
				log.Errorf("could not get wallet balance for %s: %s", wallet.Addr, err)
				continue
			}

			w[i].Balance = core.WalletBalance{
				BalanceFilecoin: bal.Balance.Balance,
				BalanceDatacap:  bal.Balance.VerifiedClientBalance,
			}
		}

		return c.JSON(200, w)
	})

	wallets.POST("", func(c echo.Context) error {
		return handleAddWallet(c, dldm)
	})

	wallets.POST("/associate", func(c echo.Context) error {
		return handleAssociateWallet(c, dldm)
	})

	wallets.DELETE("/:wallet", func(c echo.Context) error {

		w := c.Param("wallet")

		res := dldm.DB.Model(&core.Wallet{}).Where("addr = ?", w).Delete(&core.Wallet{})

		if res.Error != nil {
			return fmt.Errorf("could not delete wallet %s : %s", w, res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("wallet not found %s", w)
		}

		return c.JSON(200, "wallet successfully deleted")
	})
}

// We will not private key in DDM DB, pass it on to Delta
type PostWalletBody struct {
	Type       string `json:"Type"`
	PrivateKey string `json:"PrivateKey"`
}

type PostWalletBodyHex struct {
	HexKey string `json:"hex_key"`
}

// POST /api/wallet
// @description add/import a wallet
// @returns newly added wallet info
func handleAddWallet(c echo.Context, dldm *core.DeltaDM) error {
	authKey := c.Get(core.AUTH_KEY).(string)

	isHex := c.QueryParam("hex")

	var deltaResp *core.RegisterWalletResponse

	if isHex == "true" {
		var w PostWalletBodyHex
		var err error
		if err = c.Bind(&w); err != nil {
			return fmt.Errorf("failed to bind hex input")
		}

		deltaResp, err = dldm.DAPI.AddWalletByHexKey(core.RegisterWalletHexRequest(w), authKey)
		if err != nil {
			return fmt.Errorf("could not add wallet %s", err)
		}
		if deltaResp.WalletAddr == "" {
			return fmt.Errorf("could not add wallet, got no address back from delta. check wallet hex. delta response: %s", deltaResp.Message)
		}
	} else {
		// non-hex (priv key + type) wallet entry
		var w PostWalletBody
		var err error
		if err = c.Bind(&w); err != nil {
			return fmt.Errorf("failed to bind wallet input")
		}

		deltaResp, err = dldm.DAPI.AddWalletByPrivateKey(core.RegisterWalletRequest{
			Type:       w.Type,
			PrivateKey: w.PrivateKey,
		}, authKey)
		if err != nil {
			return fmt.Errorf("could not add wallet %s", err)
		}
		if deltaResp.WalletAddr == "" {
			return fmt.Errorf("could not add wallet, got no address back from delta. check key format and type. delta response: %s", deltaResp.Message)
		}
	}

	newWallet := core.Wallet{
		Addr: deltaResp.WalletAddr,
	}

	res := dldm.DB.Model(core.Wallet{}).Create(&newWallet)
	if res.Error != nil {
		if res.Error.Error() == "UNIQUE constraint failed: wallets.addr" {
			return fmt.Errorf("wallet %s already exists in delta", newWallet.Addr)
		}
		return res.Error
	}
	return c.JSON(200, newWallet)

}

type AssociateWalletBody struct {
	Address  string   `json:"address"`
	Datasets []string `json:"datasets"`
}

// POST /api/wallet/associate
// @description associate a wallet with a dataset
func handleAssociateWallet(c echo.Context, dldm *core.DeltaDM) error {
	var awb AssociateWalletBody

	if err := c.Bind(&awb); err != nil {
		return err
	}

	var wallet core.Wallet
	findWallet := dldm.DB.Model(core.Wallet{}).Where("addr = ?", awb.Address).Find(&wallet)
	if findWallet.Error != nil {
		return fmt.Errorf("could not find wallet %s : %s", awb.Address, findWallet.Error)
	}

	if wallet.Addr == "" {
		return fmt.Errorf("wallet %s does not exist", awb.Address)
	}

	if len(awb.Datasets) == 0 {
		return fmt.Errorf("no datasets provided")
	}

	var newDatasets []core.Dataset
	for _, datasetName := range awb.Datasets {

		var dataset core.Dataset
		err := dldm.DB.Model(core.Dataset{}).
			Where("name = ?", datasetName).
			Find(&dataset).
			Error

		if err != nil {
			return fmt.Errorf("could not check for dataset %s : %s", datasetName, err)
		}

		if dataset.ID == 0 {
			return fmt.Errorf("dataset %s does not exist", datasetName)
		}

		newDatasets = append(newDatasets, dataset)
	}

	err := dldm.DB.Model(&wallet).Association("Datasets").Replace(newDatasets)
	if err != nil {
		return fmt.Errorf("could not associate wallet with dataset: %s", err)
	}

	return c.JSON(200, "successfully associated wallet with datasets")
}
