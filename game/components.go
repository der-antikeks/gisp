package game

import (
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

const (
	GameStateType ecs.ComponentType = iota
)

type GameStateComponent struct {
	State string
	Since time.Time
}

func (c GameStateComponent) Type() ecs.ComponentType {
	return GameStateType
}
