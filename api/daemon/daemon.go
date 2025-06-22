package daemon

import (
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/server"
)

var Daemons = map[string]struct {
	Start func()
	Stop  func()
}{
	"server":   {Start: server.Start, Stop: server.Stop},
	"database": {Start: database.Start, Stop: database.Stop},
}
