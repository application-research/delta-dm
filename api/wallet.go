package api

import (
	"fmt"

	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ConfigureWalletRouter(e *echo.Group, dldm *core.DeltaDM) {
	replication := e.Group("/wallet")

	replication.GET("", func(c echo.Context) error {
		err := RequestAuthHeaderCheck(c)
		if err != nil {
			return c.JSON(401, err.Error())
		}

		authorizationString := c.Request().Header.Get("Authorization")

		ds := c.QueryParam("dataset")

		var w []core.Wallet

		tx := dldm.DB.Model(&core.Wallet{})

		if ds != "" {
			tx.Where("dataset_name = ?", ds)
		}

		tx.Find(&w)

		for i, wallet := range w {
			bal, err := dldm.DAPI.GetWalletBalance(wallet.Addr, authorizationString)
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

	replication.POST("/:dataset", func(c echo.Context) error {
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
	err := RequestAuthHeaderCheck(c)
	if err != nil {
		return c.JSON(401, err.Error())
	}

	authorizationString := c.Request().Header.Get("Authorization")

	if err := c.Bind(&w); err != nil {
		return err
	}

	ds := c.Param("dataset")

	var exists bool
	err = dldm.DB.Model(core.Dataset{}).
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

	deltaResp, err := dldm.DAPI.AddWallet(core.AddWalletRequest{
		Type:       w.Type,
		PrivateKey: w.PrivateKey,
	}, authorizationString)
	if err != nil {
		return fmt.Errorf("could not add wallet %s", err)
	}
	if deltaResp.WalletAddr == "" {
		return fmt.Errorf("could not add wallet, got no address back from delta. check key format and type. delta response: %s", deltaResp.Message)
	}

	var existingWallet core.Wallet
	res := dldm.DB.
		Model(core.Wallet{}).
		Where("dataset_name = ?", ds).
		First(&existingWallet)

	if res.Error != nil {
		if res.Error != gorm.ErrRecordNotFound {
			return res.Error
		} else {
			// No existing wallet for this dataset, create it
			newWallet := core.Wallet{
				Addr:        deltaResp.WalletAddr,
				Type:        w.Type,
				DatasetName: ds,
			}

			res = dldm.DB.Model(core.Wallet{}).Create(&newWallet)
			if res.Error != nil {
				return res.Error
			}
			return c.JSON(200, newWallet)
		}
	} else {
		// Update existing wallet
		existingWallet.Addr = deltaResp.WalletAddr
		existingWallet.Type = w.Type
		res := dldm.DB.Model(&core.Wallet{}).Where("dataset_name = ?", ds).Select("Addr", "Type").Updates(core.Wallet{Addr: deltaResp.WalletAddr, Type: w.Type})

		if res.Error != nil {
			return res.Error
		}
		return c.JSON(200, existingWallet)
	}
}
