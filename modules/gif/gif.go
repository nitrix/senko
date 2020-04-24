package gif

import (
	"fmt"
	"senko/app"
	"senko/modules/gif/tenor"
	"strings"
)

type Module struct {}

func (m Module) Dispatch(request app.Request, response app.Response) error {
	if len(request.Args) > 1 && request.Args[0] == "gif" {
		tag := strings.Join(request.Args[1:], " ")
		return m.gif(request, response, tag)
	}

	return nil
}

func (m Module) gif(request app.Request, response app.Response, tag string) error {
	tenorToken := app.GetToken("TENOR_TOKEN")
	tenorInstance := tenor.NewTenor(tenorToken)

	if request.Args[1] == "-nsfw" || request.NSFW {
		tenorInstance.NSFW = true
		tag = strings.Join(request.Args[2:], " ")
	}

	gif, err := tenorInstance.RandomGif(tag)
	if err != nil {
		return fmt.Errorf("unable to contact tenor: %w", err)
	}

	return response.SendImageFromURL(gif.URL)
}
