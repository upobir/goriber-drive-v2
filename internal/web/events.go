package web

import (
	"log"
	"time"
	"upobir/goriber-drive-v2/internal/service"
)

type EventType string

var (
	EventWelcome     EventType = "welcome"
	EventFileCreated EventType = "file.created"
	EventFileDeleted EventType = "file.deleted"
)

type WSEvent struct {
	Type      EventType `json:"type"`
	Timestamp int64     `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

func NewWSEvent(eventType EventType, data any) WSEvent {
	return WSEvent{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

type HubBroadcaster struct {
	Hub *Hub
}

func (b *HubBroadcaster) Broadcast(event service.Event) {
	switch event.Type {
	case service.EventFileCreated:
		file, ok := event.Payload.(service.File)
		if !ok {
			log.Println("failed to cast event payload to service.File")
			return
		}
		b.Hub.Broadcast(NewWSEvent(EventFileCreated, fromService(file)))
	case service.EventFileDeleted:
		id, ok := event.Payload.(string)
		if !ok {
			log.Println("failed to cast event payload to string")
			return
		}
		b.Hub.Broadcast(NewWSEvent(EventFileDeleted, id))
	}
}
