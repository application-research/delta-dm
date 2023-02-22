package api

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/application-research/delta-ldm/core"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type PostReplicationBody struct {
	Provider string  `json:"provider"`
	Dataset  *string `json:"dataset,omitempty"`
	NumDeals *uint   `json:"num_deals,omitempty"`
	// NumTib       *int    `json:"num_tib,omitempty"`
	PricePerDeal float64 `json:"price_per_deal,omitempty"`
}

func ConfigureReplicationRouter(e *echo.Group, dldm *core.DeltaLDM) {
	replication := e.Group("/replication")

	replication.GET(":provider", func(c echo.Context) error {
		var r []core.Replication
		p := c.Param("provider")

		dldm.DB.Find(&r).Where("provider_actor_id = ?", p)

		return c.JSON(200, r)
	})

	replication.POST("", func(c echo.Context) error {
		return handlePostReplication(c, dldm)
	})

}

// POST /api/replication
// @param num number of deals requested
// @returns a slice of the CIDs
func handlePostReplication(c echo.Context, dldm *core.DeltaLDM) error {
	var d PostReplicationBody

	if err := c.Bind(&d); err != nil {
		return err
	}

	if d.NumDeals == nil {
		return fmt.Errorf("must specify num_deals")
	}

	// TODO: Support num_tib to allow specifying the amount of data to replicate

	toReplicate, err := findUnreplicatedContentForProvider(dldm.DB, d.Provider, d.Dataset, d.NumDeals)
	if err != nil {
		return err
	}

	var dealsToMake core.OfflineDealRequest
	fmt.Printf("calling DELTA api for %+v deals\n\n", len(toReplicate))

	for _, c := range toReplicate {
		dealsToMake = append(dealsToMake, core.Deal{
			Cid:            c.PayloadCID, // Payload CID
			Wallet:         core.Wallet{},
			ConnectionMode: "offline",
			Miner:          d.Provider,
			Size:           c.Size,
			// TODO: duration and start epoch
			Commp: core.Commp{
				Piece:           c.CommP,
				PaddedPieceSize: c.PaddedSize,
			},
		})
	}
	deltaResp, err := dldm.DAPI.MakeOfflineDeals(dealsToMake)
	if err != nil {
		return fmt.Errorf("unable to make deal with delta api: %s", err)
	}

	for i, c := range *deltaResp {
		var newReplication = core.Replication{
			ContentCommP:    c.Meta.Cid,
			ProviderActorID: d.Provider,
			DeltaContentID:  c.ContentID,
			DealTime:        time.Now(),
			ProposalCid:     fmt.Sprint(rand.Int()) + fmt.Sprint(i), // TODO: From delta
		}

		dldm.DB.Model(&core.Replication{}).Create(&newReplication)

		for _, dbContent := range toReplicate {
			if dbContent.CommP == c.Meta.Cid {
				dbContent.NumReplications += 1
				dldm.DB.Save(&dbContent)
			}
		}
	}

	return c.JSON(200, deltaResp)
}

// Query the database for all contant that does not have replications to this actor yet
// Arguments: providerID - the actor ID of the provider
// 					  datasetName (optional) - the name of the dataset to replicate
// 					  numDeals (optional) - the number of replications (deals) to return. If nil, return all
func findUnreplicatedContentForProvider(db *gorm.DB, providerID string, datasetName *string, numDeals *uint) ([]core.Content, error) {

	rawQuery := "select * from datasets d inner join contents c " +
		"on d.id = c.dataset_id where c.comm_p not in " +
		"(select r.content_comm_p from replications r where r.provider_actor_id not in (select p.actor_id from providers p where p.actor_id not in (?))) " +
		"and c.num_replications < d.replication_quota"
	var rawValues = []interface{}{providerID}

	if datasetName != nil {
		rawQuery += " AND d.name = ?"
		rawValues = append(rawValues, datasetName)
	}

	if numDeals != nil {
		rawQuery += " LIMIT ?"
		rawValues = append(rawValues, numDeals)
	}
	var contents []core.Content
	db.Raw(rawQuery, rawValues...).Scan(&contents)

	return contents, nil
}
