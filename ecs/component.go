package ecs

// ComponentType identifies a specific Component
type ComponentType int

// Component is a set of data needed for a specific purpose
type Component interface {
	Type() ComponentType
}
