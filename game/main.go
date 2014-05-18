package game

import (
	"log"
	"math/rand"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

var (
	fps   = 70
	w, h  = 800, 400
	title = "gisp concurrent-ecs rewrite"

	initialized bool
	running     bool
)

func SetFPS(f int) {
	fps = f
}

func Stop() {
	running = false
}

func Run() error {
	// init
	rand.Seed(time.Now().Unix())
	running = true
	engine := ecs.NewEngine()

	// managers
	em := NewEntityManager(engine)
	im, wm := InitOpenGL(w, h, title, engine)
	defer wm.Cleanup()

	// systems
	NewGameStateSystem(engine, em, im, wm)
	NewMenuSystem(engine, im)
	NewRenderSystem(engine, wm)
	NewOrbitControlSystem(engine, im)

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

	for running {
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
			engine.Publish(ecs.MessageUpdate{Delta: delta})

		case <-console:
			// print fps
			log.Println(curfps)
		}
	}

	return nil
}
