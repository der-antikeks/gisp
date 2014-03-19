package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

func main() {
	rand.Seed(time.Now().Unix())
	runtime.LockOSThread()

	engine := ecs.NewEngine()

	// managers
	em := NewEntityManager(engine)
	im, wm := InitOpenGL(640, 480, "gophers in space!") // Input/WindowManager
	defer wm.cleanup()

	// systems
	engine.AddSystem(NewGameStateSystem(em, im, wm), 0)
	engine.AddSystem(NewMenuSystem(im), 5)
	engine.AddSystem(NewRenderSystem(wm), 10)

	// main loop
	var (
		lastTime    = time.Now()
		currentTime time.Time
		delta       time.Duration

		ratio     = 0.01
		fps       = 70.0
		nextPrint = lastTime

		renderTicker = time.Tick(time.Duration(1000/fps) * time.Millisecond)
	)

	for engine.IsRunning() {
		select {
		case <-renderTicker:
			// calc delay
			currentTime = time.Now()
			delta = currentTime.Sub(lastTime)
			lastTime = currentTime

			// fps test
			fps = fps*(1-ratio) + (1.0/delta.Seconds())*ratio
			if fps >= math.Inf(1) {
				fps = 72.0
			}
			if currentTime.After(nextPrint) {
				nextPrint = currentTime.Add(time.Second / 2.0)
				fmt.Println(fps)
			}

			// update
			if err := engine.Update(delta); err != nil {
				log.Fatal(err)
			}
		}
	}
}
