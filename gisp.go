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

	// managers
	em := game.NewEntityManager(engine)
	im, wm := game.InitOpenGL(w, h, title, engine)
	defer wm.Cleanup()

	// systems
	quit := make(chan struct{})
	game.NewGameStateSystem(engine, em, im, wm, quit)
	game.NewMenuSystem(engine, im)
	game.NewRenderSystem(engine, wm)
	game.NewOrbitControlSystem(engine, im)
	game.NewMovementSystem(engine)
	game.NewSceneSystem(engine)

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
			engine.Publish(game.MessageUpdate{Delta: delta})

		case <-console:
			// print fps
			log.Println(curfps)

		case <-quit:
			return
		}
	}
}
