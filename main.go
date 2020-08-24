package main

import (
	"log"
	"senko/app"
	"senko/gateways/discord"
	"senko/modules/core"
)

func main() {
	a := app.App{}

	// Gateways are responsible for generating requests and processing responses.
	a.RegisterGateway(&discord.Discord{})
	//a.RegisterGateway(&irc.IRC{})
	//a.RegisterGateway(&web.Web{})

	// Modules are responsible for processing requests and producing responses.
	//a.RegisterModule(&anime.Anime{})
	//a.RegisterModule(&autojoin.Autojoin{})
	a.RegisterModule(&core.Core{})
	//a.RegisterModule(&deejay.Deejay{})
	//a.RegisterModule(&eggplant.Eggplant{})
	//a.RegisterModule(&jarvis.Jarvis{})
	//a.RegisterModule(&youtube.Youtube{})

	// Run the application.
	err := a.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
