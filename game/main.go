package game

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

var (
	fps   = 70
	w, h  = 800, 400
	title = "gisp ecs rewrite"

	engine      *ecs.Engine
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
	runtime.LockOSThread()
	running = true
	engine = ecs.NewEngine()

	// managers
	em := NewEntityManager(engine)
	im, wm := InitOpenGL(w, h, title)
	defer wm.Cleanup()

	// systems
	engine.AddSystem(NewGameStateSystem(em, im, wm), 0)
	engine.AddSystem(NewMenuSystem(im), 5)
	engine.AddSystem(NewRenderSystem(wm), 10)

	// main loop
	var (
		lastTime = time.Now()
		now      time.Time
		delta    time.Duration
		ds       float64

		ratio  = 0.01
		curfps = float64(fps)

		update  = time.Tick(time.Duration(1000/fps) * time.Millisecond)
		console = time.Tick(500 * time.Millisecond)
	)

	for running {
		select {
		case <-update:
			// calc delay
			now = time.Now()
			delta = now.Sub(lastTime)
			lastTime = now

			// calc fps
			if ds = delta.Seconds(); ds > 0 {
				curfps = curfps*(1-ratio) + (1.0/ds)*ratio
			}

			// update
			if err := engine.Update(delta); err != nil {
				return err
			}
		case <-console:
			// print fps
			log.Println(curfps)
		}
	}

	return nil
}
