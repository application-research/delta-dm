package api

import (
	"fmt"
	"net/http"

	"github.com/application-research/delta-dm/core"
	db "github.com/application-research/delta-dm/db"
	"github.com/labstack/echo/v4"
)

type ReplicationProfile struct {
	DatasetName string `json:"dataset_name"`
	Unsealed    bool   `json:"unsealed"`
	Indexed     bool   `json:"indexed"`
}

func ConfigureReplicationProfilesRouter(e *echo.Group, dldm *core.DeltaDM) {
	replicationProfiles := e.Group("/replication-profiles")

	replicationProfiles.Use(dldm.AS.AuthMiddleware)

	replicationProfiles.GET("", func(c echo.Context) error {
		var p []db.ReplicationProfile

		res := dldm.DB.Model(&db.ReplicationProfile{}).Find(&p)
		if res.Error != nil {
			return fmt.Errorf("error finding replication profiles: %s", res.Error)
		}

		return c.JSON(http.StatusOK, p)
	})

	replicationProfiles.POST("", func(c echo.Context) error {
		var p db.ReplicationProfile

		if err := c.Bind(&p); err != nil {
			return fmt.Errorf("failed to parse request body: %s", err.Error())
		}

		// Check if the dataset and provider exist
		var ds db.Dataset
		var provider db.Provider

		dsRes := dldm.DB.Where("id = ?", p.DatasetID).First(&ds)
		providerRes := dldm.DB.Where("actor_id = ?", p.ProviderActorID).First(&provider)

		if dsRes.Error != nil || ds.ID == 0 {
			return fmt.Errorf("invalid dataset: %s", dsRes.Error)
		}

		if providerRes.Error != nil || provider.ActorID == "" {
			return fmt.Errorf("invalid provider: %s", providerRes.Error)
		}

		// Save the replication profile
		res := dldm.DB.Create(&p)
		if res.Error != nil {
			if res.Error.Error() == "UNIQUE constraint failed: replication_profiles.provider_actor_id, replication_profiles.dataset_id" {
				return fmt.Errorf("replication profile for provider %s and datasetID %d already exists", p.ProviderActorID, p.DatasetID)
			}
			return fmt.Errorf("failed to save replication profile: %s", res.Error.Error())
		}

		return c.JSON(http.StatusOK, p)
	})

	replicationProfiles.DELETE("", func(c echo.Context) error {
		var p db.ReplicationProfile
		if err := c.Bind(&p); err != nil {
			return fmt.Errorf("failed to parse request body: %s", err.Error())
		}

		if p.DatasetID == 0 || p.ProviderActorID == "" {
			return fmt.Errorf("invalid replication profile ID")
		}

		var existingProfile db.ReplicationProfile
		res := dldm.DB.Where("provider_actor_id = ? AND dataset_id = ?", p.ProviderActorID, p.DatasetID).First(&existingProfile)

		if res.Error != nil {
			return fmt.Errorf("replication profile not found: %s", res.Error)
		}

		deleteRes := dldm.DB.Delete(&existingProfile)
		if deleteRes.Error != nil {
			return fmt.Errorf("failed to delete replication profile: %s", deleteRes.Error.Error())
		}

		return c.JSON(http.StatusOK, fmt.Sprintf("replication profile with ProviderActorID %s and DatasetID %d deleted successfully", p.ProviderActorID, p.DatasetID))
	})

	replicationProfiles.PUT("", func(c echo.Context) error {
		var updatedProfile db.ReplicationProfile
		if err := c.Bind(&updatedProfile); err != nil {
			return fmt.Errorf("failed to parse request body: %s", err.Error())
		}

		if updatedProfile.DatasetID == 0 || updatedProfile.ProviderActorID == "" {
			return fmt.Errorf("invalid replication profile ID")
		}

		var existingProfile db.ReplicationProfile
		res := dldm.DB.Where("provider_actor_id = ? AND dataset_id = ?", updatedProfile.ProviderActorID, updatedProfile.DatasetID).First(&existingProfile)

		if res.Error != nil {
			return fmt.Errorf("replication profile not found: %s", res.Error)
		}

		updateData := map[string]interface{}{
			"unsealed": updatedProfile.Unsealed,
			"indexed":  updatedProfile.Indexed,
		}

		updateRes := dldm.DB.Model(&existingProfile).Updates(updateData)
		if updateRes.Error != nil {
			return fmt.Errorf("failed to update replication profile: %s", updateRes.Error.Error())
		}

		return c.JSON(http.StatusOK, updatedProfile)
	})

}
