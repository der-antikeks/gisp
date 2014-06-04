package game

import (
	"time"
)

type MessageEntityAdd struct{ Added Entity }
type MessageEntityRemove struct{ Removed Entity }
type MessageEntityUpdate struct{ Updated Entity }

type MessageUpdate struct{ Delta time.Duration }

type MessageKey Key
type MessageTimeout time.Time

type MessagePause struct{}
type MessageContinue struct{}
type MessageQuit struct{}

type MessageMouseButton MouseButton
type MessageMouseMove struct{ X, Y float64 }
type MessageResize struct{ Width, Height int }
type MessageMouseScroll float64
