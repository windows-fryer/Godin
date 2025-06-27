package service

import (
	"database/sql"
	"net/http"

	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/resource"
)

const (
	GetServiceQuery = "SELECT service_id FROM eris.services WHERE guild_id = $1"
)

type GetServiceResponse struct {
	ServiceID string `json:"service_id"`
}

func getService(tx *sql.Tx, guildID string) (string, error) {
	var serviceID string

	if err := tx.QueryRow(GetServiceQuery, guildID).Scan(&serviceID); err != nil {
		return "", err
	}

	return serviceID, nil
}

func (service *DiscordAPIService) GetService(w http.ResponseWriter, r *http.Request) error {
	guildID := r.URL.String()[len("/v1/service/eris/"):]

	tx, err := database.PostgresClient.BeginTx(r.Context(), nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	serviceID, err := getService(tx, guildID)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	response := GetServiceResponse{
		ServiceID: serviceID,
	}

	return resource.GenerateJSONResponse(w, http.StatusOK, response)
}
