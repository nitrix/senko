package main

import (
	"log"
	"senko/app"
	"senko/modules/autojoin"
	"senko/modules/core"
	"senko/modules/deejay"
	"senko/modules/eggplant"
	"senko/modules/jarvis"
)

func main() {
	a := app.App{}

	// a.RegisterModule(&anime.Anime{})
	a.RegisterModule(&autojoin.Autojoin{})
	a.RegisterModule(&core.Core{})
	a.RegisterModule(&deejay.Deejay{})
	a.RegisterModule(&eggplant.Eggplant{})
	a.RegisterModule(&jarvis.Jarvis{})

	// Run the application.
	err := a.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
