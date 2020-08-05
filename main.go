package main

import (
	"senko/app"
	"senko/module/anime"
	"senko/module/autojoin"
	"senko/module/core"
	"senko/module/deejay"
	"senko/module/eggplant"
	"senko/module/jarvis"
	"senko/module/youtube"
)

func main() {
	a := app.App{}

	a.RegisterModule(&anime.Anime{})
	a.RegisterModule(&autojoin.Autojoin{})
	a.RegisterModule(&core.Core{})
	a.RegisterModule(&deejay.Deejay{})
	a.RegisterModule(&eggplant.Eggplant{})
	a.RegisterModule(&jarvis.Jarvis{})
	a.RegisterModule(&youtube.Youtube{})

	a.Run()
}
