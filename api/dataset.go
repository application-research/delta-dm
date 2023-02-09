package api

import "github.com/labstack/echo/v4"

func ConfigureDatasetRouter(e *echo.Group) {
	dataset := e.Group("/dataset")

	dataset.POST("/add", func(c echo.Context) error {
		return nil
	})
}
