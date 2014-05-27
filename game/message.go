package ecs

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

	UserMessages
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
