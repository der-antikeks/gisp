package ecs

import (
	"time"
)

type Event interface{}

type EntityAddEvent struct {
	Event
	Added *Entity
}

type EntityRemoveEvent struct {
	Event
	Removed *Entity
}

type EntityUpdateEvent struct {
	Event
	Updated *Entity
}

type UpdateEvent struct {
	Event
	Delta time.Duration
}

type ShutdownEvent struct {
	Event
}
