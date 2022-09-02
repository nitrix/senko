package main

import (
	"log"

	"github.com/nitrix/senko/modules"
)

func main() {
	app := App{}

	app.RegisterModule(&modules.Autojoin{})
	app.RegisterModule(&modules.Core{})
	app.RegisterModule(&modules.Justin{})

	err := app.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
