package api

import (
	"fmt"
	"time"

	"github.com/application-research/delta-ldm/core"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MakeDealBody struct {
	Provider string `json:"provider"`
	// Dataset      string  `json:"dataset"`
	NumDeals     *uint   `json:"num_deals,omitempty"`
	NumTib       *int    `json:"num_tib,omitempty"`
	PricePerDeal float64 `json:"price_per_deal,omitempty"`
}

func ConfigureDealRouter(e *echo.Group, db *gorm.DB) {
	providers := e.Group("/deal")

	providers.GET("", func(c echo.Context) error {
		var p []core.Provider

		db.Find(&p)

		return c.JSON(200, p)
	})

	providers.POST("", func(c echo.Context) error {
		return HandlePostDeal(c, db)
	})

}

// TODO: Rate limit this API per user
// Only allow max of 1 request per minute per actorID

// POST /api/deal
// @param num number of deals requested
// @returns a slice of the CIDs
func HandlePostDeal(c echo.Context, db *gorm.DB) error {
	var d MakeDealBody
	// var dataset core.Dataset

	if err := c.Bind(&d); err != nil {
		return err
	}

	if d.NumDeals == nil && d.NumTib == nil {
		return fmt.Errorf("must specify either num_deals or num_tib")
	}

	// err := db.Where("name = ?", d.Dataset).First(&dataset)
	// if err != nil {
	// 	return err.Error
	// }

	if d.NumDeals != nil {
		toReplicate, err := findUnreplicatedContentForProvider(db, d.Provider, *d.NumDeals)
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
				ProposalCid:     "aaa" + fmt.Sprint(i), // TODO: From delta
			}
			db.Model(&core.Replication{}).Create(&newReplication)
			c.NumReplications += 1
			db.Save(&c)
		}

	} else {
		// TODO: make the number of tib provided
	}

	return nil
}

// Query the database for all contant that does not have replications to this actor yet
// find count # of them and return them
func findUnreplicatedContentForProvider(db *gorm.DB, providerID string, numDeals uint) ([]core.Content, error) {
	// var dataset core.Dataset
	var contents []core.Content
	db.Raw("select * from datasets d inner join contents c "+
		"on d.id = c.dataset_id where c.comm_p not in "+
		"(select r.content_comm_p from replications r where r.provider_actor_id not in (select p.actor_id from providers p where p.actor_id not in (?))) "+
		"and c.num_replications < d.replication_quota "+
		"LIMIT ?", providerID, numDeals).Scan(&contents)

	return contents, nil
}
