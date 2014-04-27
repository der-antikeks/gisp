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

		ratio     = 0.01
		curfps    = float64(fps)
		nextPrint = lastTime

		ticker = time.Tick(time.Duration(1000/fps) * time.Millisecond)
	)

	for running {
		select {
		case now = <-ticker:
			// calc delay
			delta = now.Sub(lastTime)
			lastTime = now

			// fps test
			if ds = delta.Seconds(); ds > 0 {
				curfps = curfps*(1-ratio) + (1.0/ds)*ratio
			}
			if now.After(nextPrint) {
				nextPrint = now.Add(time.Second / 2.0)
				log.Println(curfps)
			}

			// update
			if err := engine.Update(delta); err != nil {
				return err
			}
		}
	}

	return nil
}
