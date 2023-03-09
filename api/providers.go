package api

import (
	"net/http"

	"github.com/application-research/delta-dm/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func ConfigureProvidersRouter(e *echo.Group, dldm *core.DeltaDM) {
	providers := e.Group("/providers")

	providers.GET("", func(c echo.Context) error {
		var p []core.Provider

		dldm.DB.Find(&p)

		for i, sp := range p {
			var rb [2]uint64
			dldm.DB.Raw("select SUM(size) s, SUM(padded_size) ps FROM contents c inner join replications r on r.content_comm_p = c.comm_p where r.status = 'SUCCESS' AND r.provider_actor_id = ?", sp.ActorID).Row().Scan(&rb[0], &rb[1])

			p[i].ReplicatedBytes = core.ByteSizes{Raw: rb[0], Padded: rb[1]}
		}

		return c.JSON(200, p)
	})

	providers.POST("", func(c echo.Context) error {
		var p core.Provider

		if err := c.Bind(&p); err != nil {
			return err
		}

		p.Key = uuid.New()

		res := dldm.DB.Create(&p)

		if res.Error != nil {
			return res.Error
		}
		return c.JSON(http.StatusOK, p)
	})

}
