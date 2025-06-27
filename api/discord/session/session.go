package session

import (
	"time"

	"github.com/golang/glog"
	"wednesday.wtf/godin/internal/database"
)

const (
	DeleteExpiredSessionsQuery = "DELETE FROM eris.sessions WHERE expiration_time < NOW() "
)

type DiscordAPISession struct{}

func ClearExpiredSessions() {
	timer := time.NewTicker(1 * time.Minute)

	go func() {
		for {
			<-timer.C

			if _, err := database.PostgresClient.Exec(DeleteExpiredSessionsQuery); err != nil {
				glog.Fatalf("Error clearing expired sessions: %v", err)
			}
		}
	}()
}
