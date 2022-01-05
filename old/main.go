package main

import (
	"log"
	"senko/app"
	"senko/modules/alias"
	"senko/modules/autojoin"
	"senko/modules/core"
	"senko/modules/dectalk"
	"senko/modules/deejay"
	"senko/modules/eggplant"
	"senko/modules/jarvis"
	"senko/modules/reinstall"
	"senko/modules/smart"
)

func main() {
	a := app.App{}

	a.RegisterModule(&alias.Alias{})
	//a.RegisterModule(&anime.Anime{})
	a.RegisterModule(&autojoin.Autojoin{})
	a.RegisterModule(&core.Core{})
	a.RegisterModule(&dectalk.Dectalk{})
	a.RegisterModule(&deejay.Deejay{})
	a.RegisterModule(&eggplant.Eggplant{})
	a.RegisterModule(&reinstall.Reinstall{})
	a.RegisterModule(&jarvis.Jarvis{})
	a.RegisterModule(&smart.Smart{})

	// Run the application.
	err := a.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
