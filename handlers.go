package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/filecoin-project/go-address"
	"github.com/labstack/echo/v4"
)

// TODO: Rate limit this API per user
// Only allow max of 1 request per minute per actorID

// POST /api/deal
// @param key the API key / password to request deals
// @param num number of deals requested
// @returns a slice of the CIDs
func (t *DeltaLDM) HandlePostDeal(c echo.Context) error {
	key, err := address.NewFromString(c.Param("key"))
	if err != nil {
		return err
	}

	fmt.Println(key)

	num, err := strconv.ParseUint(c.Param("num"), 10, 32)
	if err != nil {
		return err
	}

	if num > 32 {
		return fmt.Errorf("maximum deal request limit is 32")
	}

	// look up the key in the DB, map it back to the SP
	// from that, you will have an Actor ID
	actor := "f012345"

	// Look up ALL CIDs that have not been assigned yet to this actor,
	// pick them at random (from any slug), and grab the CAR file as well as the associated datacap wallet address
	// Use that to make the deal

	// Make boost deals for CIDs that have not been assigned yet
	// For each that is successful, mark the CID in the DB that it has been dealt

	// t.deal_length
	// t.piece_size

	// cids := ["Qm12345", "Qm64213"]

	return c.JSON(http.StatusOK, actor)
}
