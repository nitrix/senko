package main

import (
	"senko/app"
	"senko/modules/anime"
	"senko/modules/gif"
	"senko/modules/orphan"
	"senko/modules/youtube"
)

func main() {
	app.RegisterModule(anime.Module{})
	app.RegisterModule(gif.Module{})
	app.RegisterModule(orphan.Module{})
	app.RegisterModule(youtube.Module{})
	app.Run()
}
