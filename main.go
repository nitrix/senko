package main

import (
	"senko/app"
	"senko/module/anime"
	"senko/module/core"
	"senko/module/eggplant"
	"senko/module/experimental"
	"senko/module/youtube"
)

func main() {
	a := app.App{}

	a.RegisterModule(&anime.Anime{})
	a.RegisterModule(&core.Core{})
	a.RegisterModule(&eggplant.Eggplant{})
	a.RegisterModule(&youtube.Youtube{})
	a.RegisterModule(&experimental.Experimental{})

	a.Run()
}
