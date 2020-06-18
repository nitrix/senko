package main

import (
	"senko/app"
	"senko/plugins/anime"
	"senko/plugins/core"
	"senko/plugins/eggplant"
	"senko/plugins/experimental"
	"senko/plugins/youtube"
)

func main() {
	app.RegisterPlugin(&anime.Plugin{})
	app.RegisterPlugin(&core.Plugin{})
	app.RegisterPlugin(&eggplant.Plugin{})
	app.RegisterPlugin(&youtube.Plugin{})
	app.RegisterPlugin(&experimental.Plugin{})
	app.Run()
}
