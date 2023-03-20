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

		tx := dldm.DB.Model(&core.Wallet{})

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

	ds := c.QueryParam("dataset")
	isHex := c.QueryParam("hex")

	// Pre-check: ensure dataset exists before doing anything
	if ds != "" {
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
	}

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

	if ds != "" {
		newWallet.DatasetName = ds
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
	Address string `json:"address"`
	Dataset string `json:"dataset"`
}

// POST /api/wallet/associate
// @description associate a wallet with a dataset
func handleAssociateWallet(c echo.Context, dldm *core.DeltaDM) error {
	var awb AssociateWalletBody

	if err := c.Bind(&awb); err != nil {
		return err
	}

	var exists bool
	err := dldm.DB.Model(core.Dataset{}).
		Select("count(*) > 0").
		Where("name = ?", awb.Dataset).
		Find(&exists).
		Error

	if err != nil {
		return fmt.Errorf("could not check if dataset %s exists: %s", awb.Dataset, err)
	}

	if !exists {
		return fmt.Errorf("dataset %s does not exist", awb.Dataset)
	}

	res := dldm.DB.Model(&core.Wallet{}).Find(&core.Wallet{Addr: awb.Address}).Update("dataset_name", awb.Dataset)

	if res.RowsAffected == 0 {
		return fmt.Errorf("wallet not found %s", awb.Address)
	}

	if res.Error != nil {
		return fmt.Errorf("could not associate wallet with dataset: %s", res.Error)
	}

	return c.JSON(200, "successfully associated wallet")
}
