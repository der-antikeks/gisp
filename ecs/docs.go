/*
	Entity component system
	http://en.wikipedia.org/wiki/Entity_component_system

	The entity is a general purpose object. It only consists of a unique id and a set of components.

	The component consists of a minimal set of data needed for a specific purpose. It may contain a behaviour associated with that data.

	The System is a single purpose function that takes a set of entities which have a specific set of components and updates them.
*/
package ecs
