package orphan

import (
	"senko/app"
)

type Module struct {}

func (m Module) Dispatch(request app.Request, response app.Response) error {
	if len(request.Args) == 1 && request.Args[0] == "version" {
		return m.version(response)
	}

	if len(request.Args) == 1 && request.Args[0] == "help" {
		return m.help(response)
	}

	return nil
}

func (m Module) version(response app.Response) error {
	return response.SendText(app.Version)
}

func (m Module) help(response app.Response) error {
	return response.SendText("For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
}