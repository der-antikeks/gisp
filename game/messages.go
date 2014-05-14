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