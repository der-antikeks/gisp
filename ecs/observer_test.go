package ecs

import (
	"fmt"
	"testing"
)

func TestCollectionObserver(t *testing.T) {
	obs := new(EntityObserver)
	var cnt int

	obs.Subscribe(func(en *Entity) {
		cnt--
	})

	obs.Subscribe(func(en *Entity) {
		cnt--
	})

	n := 10
	for i := 0; i < n; i++ {
		cnt += 2
		obs.Publish(NewEntity(fmt.Sprintf("Entity %v", i)))
	}

	if cnt != 0 {
		t.Errorf("missed a notification, remaining: %v", cnt)
	}
}
