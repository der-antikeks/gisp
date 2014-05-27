package game

import ()

type SystemPriority int

const (
	PriorityBeforeRender SystemPriority = iota
	PriorityRender
	PriorityAfterRender
)
