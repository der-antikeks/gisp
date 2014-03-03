package ecs

import ()

type EntityObserver []func(en *Entity)

// TODO: sync.RWMutex

func (obs *EntityObserver) Subscribe(f func(en *Entity)) {
	*obs = append(*obs, f)
}

func (obs *EntityObserver) Unsubscribe(f func(en *Entity)) {
	// TODO: func can only be compared to nil
	/*
		for _, o := range *obs {
			if o == f {
			}
		}
	*/
}

func (obs *EntityObserver) Publish(en *Entity) {
	for _, o := range *obs {
		o(en)
	}
}
