package main

import (
	"log"

	"github.com/der-antikeks/gisp/game"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if err := game.Run(); err != nil {
		log.Fatal(err)
	}
}
