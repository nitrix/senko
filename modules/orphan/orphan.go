package orphan

import (
	"fmt"
	"senko/app"
)

type Module struct {}

func (m Module) Dispatch(request app.Request, response app.Response) error {
	if len(request.Args) == 1 && request.Args[0] == "version" {
		return m.Version(response)
	}

	if len(request.Args) == 1 && request.Args[0] == "help" {
		return m.Help(response)
	}

	return nil
}

func (m Module) Version(response app.Response) error {
	_, err := response.Session.ChannelMessageSend(response.ChannelId, app.Version)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

func (m Module) Help(response app.Response) error {
	_, err := response.Session.ChannelMessageSend(response.ChannelId, "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}