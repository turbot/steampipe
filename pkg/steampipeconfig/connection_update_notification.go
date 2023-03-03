package steampipeconfig

import (
	"golang.org/x/exp/maps"
)

type ConnectionUpdateNotification struct {
	Update []string
	Delete []string
}

func NewConnectionUpdateNotification(updates *ConnectionUpdates) *ConnectionUpdateNotification {
	return &ConnectionUpdateNotification{
		Update: maps.Keys(updates.Update),
		Delete: maps.Keys(updates.Delete),
	}
}
