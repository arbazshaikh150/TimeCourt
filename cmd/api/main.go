package main

import (
	"log"

	"github.com/arbazshaikh150/TimeCourt/cmd/app"
)

func main() {
	app := app.NewApp()
	if err := app.Run(); err != nil {
		log.Fatal("application stopped:", err)
	}
}
