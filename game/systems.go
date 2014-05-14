package game

import (
	"github.com/der-antikeks/gisp/ecs"
)

const (
	PriorityBeforeRender ecs.SystemPriority = iota
	PriorityRender
	PriorityAfterRender
)
