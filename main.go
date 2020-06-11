package main

import (
	"senko/app"
	"senko/plugins/anime"
	"senko/plugins/core"
	"senko/plugins/eggplant"
	"senko/plugins/youtube"
)

func main() {
	app.RegisterPlugin(&anime.Plugin{})
	app.RegisterPlugin(&core.Plugin{})
	app.RegisterPlugin(&eggplant.Plugin{})
	app.RegisterPlugin(&youtube.Plugin{})
	app.Run()
}
