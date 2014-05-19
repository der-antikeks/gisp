package game

import (
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

const (
	_ ecs.MessageType = ecs.UserMessages + iota

	KeyMessageType
	MouseButtonMessageType
	MouseMoveMessageType
	MouseScrollMessageType
	ResizeMessageType
	TimeoutMessageType

	PauseMessageType
	ContinueMessageType
)

type MessageKey Key

func (e MessageKey) Type() ecs.MessageType { return KeyMessageType }

type MessageTimeout time.Time

func (e MessageTimeout) Type() ecs.MessageType { return TimeoutMessageType }

type MessagePause struct{}

func (e MessagePause) Type() ecs.MessageType { return PauseMessageType }

type MessageContinue struct{}

func (e MessageContinue) Type() ecs.MessageType { return ContinueMessageType }

type MessageMouseButton MouseButton

func (e MessageMouseButton) Type() ecs.MessageType { return MouseButtonMessageType }

type MessageMouseMove struct {
	X, Y float64
}

func (e MessageMouseMove) Type() ecs.MessageType { return MouseMoveMessageType }

type MessageResize struct {
	Width, Height int
}

func (e MessageResize) Type() ecs.MessageType { return ResizeMessageType }

type MessageMouseScroll float64

func (e MessageMouseScroll) Type() ecs.MessageType { return MouseScrollMessageType }
