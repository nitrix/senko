package main

import (
	"log"
	"os"

	"github.com/nitrix/senko/modules"
)

func main() {
	app := App{}

	app.RegisterModule(&modules.Core{})
	app.RegisterModule(&modules.Justin{})

	err := app.Run()
	if err != nil {
		log.Fatalf("Error: %s\n", err)
		os.Exit(1)
	}
}
