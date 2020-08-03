package main

import (
	"senko/app"
	"senko/module/anime"
	"senko/module/autojoin"
	"senko/module/core"
	"senko/module/deejay"
	"senko/module/eggplant"
	"senko/module/youtube"
	//"senko/module/experimental"
)

func main() {
	a := app.App{}

	a.RegisterModule(&anime.Anime{})
	a.RegisterModule(&autojoin.Autojoin{})
	a.RegisterModule(&core.Core{})
	a.RegisterModule(&deejay.Deejay{})
	a.RegisterModule(&eggplant.Eggplant{})
	a.RegisterModule(&youtube.Youtube{})
	//a.RegisterModule(&experimental.Experimental{})

	a.Run()
}
