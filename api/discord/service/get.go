package service

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/golang/glog"
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

	err := tx.QueryRow(GetServiceQuery, guildID).Scan(&serviceID)

	if err != nil {
		return "", err
	}

	return serviceID, nil
}

func (service *DiscordAPIService) GetService(w http.ResponseWriter, r *http.Request) error {
	guildID := r.URL.String()[len("/v1/service/eris/"):]

	glog.V(2).Infof("GetService endpoint used for guild_id: %s", guildID)

	if guildID == "" {
		return errors.New("guild_id is required")
	}

	tx, err := database.PostgresClient.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	serviceID, err := getService(tx, guildID)

	if err != nil {
		return err
	}

	response := GetServiceResponse{
		ServiceID: serviceID,
	}

	if err := resource.GenerateJSONResponse(w, http.StatusOK, response); err != nil {
		return err
	}

	return nil
}
