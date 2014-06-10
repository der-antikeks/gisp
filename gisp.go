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

	context := game.NewGlContextSystem(title, w, h)
	defer context.Cleanup()

	loader := game.NewAssetLoaderSystem("assets/", context)
	defer loader.Cleanup()

	ents := game.NewEntitySystem(loader)
	state := game.NewGameStateSystem(context, ents)
	spatial := game.NewSpatialSystem(ents)
	game.NewRenderSystem(context, spatial, state, ents /*temporary*/)

	game.NewMovementsSystem(ents, state)
	game.NewControlSystem(context, ents, state)

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
		quit    = make(chan interface{})
	)

	state.OnQuit().Subscribe(quit, game.PriorityLast)

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
			state.Update(delta)

		case <-console:
			// print fps
			log.Println(curfps)

		case <-quit:
			return
		}
	}
}
