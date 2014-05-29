package game

import (
	"time"
)

type MessageType int

const (
	_ MessageType = iota
	UpdateMessageType
	EntityAddMessageType
	EntityRemoveMessageType
	EntityUpdateMessageType

	KeyMessageType
	MouseButtonMessageType
	MouseMoveMessageType
	MouseScrollMessageType
	ResizeMessageType
	TimeoutMessageType

	PauseMessageType
	ContinueMessageType
	QuitMessageType
)

type Message interface {
	Type() MessageType
}

type EntityMessage interface {
	Message
	Entity() Entity
}

type MessageEntityAdd struct {
	EntityMessage
	Added Entity
}

func (e MessageEntityAdd) Type() MessageType { return EntityAddMessageType }
func (e MessageEntityAdd) Entity() Entity    { return e.Added }

type MessageEntityRemove struct {
	EntityMessage
	Removed Entity
}

func (e MessageEntityRemove) Type() MessageType { return EntityRemoveMessageType }
func (e MessageEntityRemove) Entity() Entity    { return e.Removed }

type MessageEntityUpdate struct {
	EntityMessage
	Updated Entity
}

func (e MessageEntityUpdate) Type() MessageType { return EntityUpdateMessageType }
func (e MessageEntityUpdate) Entity() Entity    { return e.Updated }

type MessageUpdate struct {
	Message
	Delta time.Duration
}

func (e MessageUpdate) Type() MessageType { return UpdateMessageType }

type Filter struct {
	Types  []MessageType
	Aspect []ComponentType
}

type MessageKey Key

func (e MessageKey) Type() MessageType { return KeyMessageType }

type MessageTimeout time.Time

func (e MessageTimeout) Type() MessageType { return TimeoutMessageType }

type MessagePause struct{}

func (e MessagePause) Type() MessageType { return PauseMessageType }

type MessageContinue struct{}

func (e MessageContinue) Type() MessageType { return ContinueMessageType }

type MessageQuit struct{}

func (e MessageQuit) Type() MessageType { return QuitMessageType }

type MessageMouseButton MouseButton

func (e MessageMouseButton) Type() MessageType { return MouseButtonMessageType }

type MessageMouseMove struct {
	X, Y float64
}

func (e MessageMouseMove) Type() MessageType { return MouseMoveMessageType }

type MessageResize struct {
	Width, Height int
}

func (e MessageResize) Type() MessageType { return ResizeMessageType }

type MessageMouseScroll float64

func (e MessageMouseScroll) Type() MessageType { return MouseScrollMessageType }
