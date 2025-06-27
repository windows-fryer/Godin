package service

import (
	"net/http"

	"wednesday.wtf/godin/internal/database"
)

const (
	DeleteServiceQuery = "DELETE FROM eris.guilds WHERE guild_id = $1"
)

func (service *DiscordAPIService) DeleteService(w http.ResponseWriter, r *http.Request) error {
	serviceID := r.URL.String()[len("/v1/service/eris/"):]

	tx, err := database.PostgresClient.BeginTx(r.Context(), nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.Exec(DeleteServiceQuery, serviceID); err != nil {
		return err
	}

	return tx.Commit()
}
