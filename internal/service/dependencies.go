package service

import (
	"database/sql"
)

type EventType string

var (
	EventFileCreated EventType = "file.created"
	EventFileDeleted EventType = "file.deleted"
)

type Event struct {
	Type    EventType
	Payload any
}

type Broadcaster interface {
	Broadcast(event Event)
}

type Dependencies struct {
	StorageDir  string
	Db          *sql.DB
	Broadcaster Broadcaster
}
