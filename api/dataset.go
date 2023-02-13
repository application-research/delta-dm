package api

import "github.com/labstack/echo/v4"

type AddDatasetBody struct {
	Name             string `json:"name"`
	ReplicationQuota int    `json:"replication_quota"`
	DealDuration     int    `json:"deal_duration"`
	Wallet           string `json:"wallet"`
	Unsealed         bool   `json:"unsealed"`
	Indexed          bool   `json:"indexed"`
}

func ConfigureDatasetRouter(e *echo.Group) {
	dataset := e.Group("/dataset")

	dataset.POST("/add", func(c echo.Context) error {
		var ads AddDatasetBody

		if err := c.Bind(&ads); err != nil {
			return err
		}

		return c.JSON(200, ads)
	})
}
