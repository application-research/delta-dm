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
		return handlePostReplication(c, dldm.DB)
	})

}

// POST /api/replication
// @param num number of deals requested
// @returns a slice of the CIDs
func handlePostReplication(c echo.Context, db *gorm.DB) error {
	var d PostReplicationBody

	if err := c.Bind(&d); err != nil {
		return err
	}

	if d.NumDeals == nil {
		return fmt.Errorf("must specify num_deals")
	}

	// TODO: Support num_tib to allow specifying the amount of data to replicate

	toReplicate, err := findUnreplicatedContentForProvider(db, d.Provider, d.Dataset, d.NumDeals)
	if err != nil {
		return err
	}

	// Deal successfully made
	for i, c := range toReplicate {
		// TODO: make the deals
		// CALL delta API
		fmt.Printf("calling DELTA api for %+v\n\n", c)
		// error check - if it fails, then don't update the DB and return an error

		var newReplication = core.Replication{
			ContentCommP:    c.CommP,
			ProviderActorID: d.Provider,
			DealTime:        time.Now(),
			ProposalCid:     fmt.Sprint(rand.Int()) + fmt.Sprint(i), // TODO: From delta
		}
		db.Model(&core.Replication{}).Create(&newReplication)
		c.NumReplications += 1
		db.Save(&c)
	}

	return c.JSON(200, toReplicate)
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
