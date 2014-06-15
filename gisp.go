package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/der-antikeks/gisp/game"
)

func main() {
	// init
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().Unix())

	context := game.GlContextSystem(&game.CtxOpts{
		Title: "gisp ecs",
		W:     800,
		H:     600,
	})
	defer context.Cleanup()

	loader := game.AssetLoaderSystem(&game.AssetOpts{Path: "assets/"})
	defer loader.Cleanup()

	state := game.GameStateSystem()

	// TODO: start/stop from gamestate-system?
	game.RenderSystem()
	game.MovementSystem()
	game.ControlSystem()

	// main loop
	var (
		fps    = 70
		ratio  = 0.01
		curfps = float64(fps)

		lastTime    = time.Now()
		currentTime time.Time
		delta       time.Duration
		ds          float64

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
