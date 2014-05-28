package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/der-antikeks/gisp/game"
)

var (
	fps   = 70
	w, h  = 800, 400
	title = "gisp engine-less-ecs rewrite"
)

func main() {
	// init
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().Unix())
	engine := game.NewEngine()

	// new systems

	// swap buffers, poll events, manage window
	context := NewGlContextSystem(title, w, h)
	defer context.Cleanup()

	// geometry, material, texture, shader
	loader := NewAssetLoaderSystem("/assets/")
	defer loader.Cleanup()

	// create, load/save entities, manage components
	ent := NewEntitySystem(loader)

	// collisions, visibility of spatially aware entities
	spatial := NewSpatialSystem(ent)

	// move entities with velocity
	move := NewMovementSystem(ent)

	// change entities based on controller input
	ctrl := NewControlSystem(context, ent)

	// manage render passes, priorities and render to screen/buffer
	render := NewRenderSystem(context, ent)

	// handle game-state, start loading/unloading entities, send update messages
	state := NewGameStateSystem(context, ent, render)
	quit := make(chan game.Message)
	state.OnQuit().Subscribe(quit, game.PriorityLast)

	// main loop
	var (
		lastTime    = time.Now()
		currentTime time.Time
		delta       time.Duration
		ds          float64

		ratio  = 0.01
		curfps = float64(fps)

		update  = time.Tick(time.Duration(1000/fps) * time.Millisecond)
		console = time.Tick(500 * time.Millisecond)
	)

	for {
		select {
		case <-update:
			// calc delay
			currentTime = time.Now()
			delta = currentTime.Sub(lastTime)
			lastTime = currentTime

			// calc fps
			if ds = delta.Seconds(); ds > 0 {
				curfps = curfps*(1-ratio) + (1.0/ds)*ratio
			}

			// update
			state.OnUpdate().Publish(game.MessageUpdate{Delta: delta})

		case <-console:
			// print fps
			log.Println(curfps)

		case <-quit:
			return
		}
	}
}
