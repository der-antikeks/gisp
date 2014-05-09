package ecs

// General purpose object that identifies a set of components.
type Entity int

// ComponentType identifies a specific Component
type ComponentType int

/*
type Property struct {
	Name  string
	Value interface{}
}
*/

// Component is a set of data needed for a specific purpose
type Component interface {
	Type() ComponentType
	//Load(<-chan Property) error
	//Save(chan<- Property) error
}
