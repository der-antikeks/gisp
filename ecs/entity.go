package ecs

import ()

type EntityList interface {
	//Add(*Entity)
	//Remove(*Entity)
	Entities() []Entity
	First() Entity
}

type SliceEntityList []Entity

func (l *SliceEntityList) Add(e Entity) {
	*l = append(*l, e)
}
func (l *SliceEntityList) Remove(e Entity) {
	a := *l
	for i, f := range a {
		if f == e {
			*l = append(a[:i], a[i+1:]...)
			return
		}
	}
}
func (l SliceEntityList) Entities() []Entity {
	return l
}
func (l SliceEntityList) First() (Entity, bool) {
	if len(l) < 1 {
		return 0, false
	}
	return l[0], true
}
