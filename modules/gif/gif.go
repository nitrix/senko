package gif

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"senko/app"
	"senko/modules/gif/tenor"
	"strings"
)

type Module struct {}

func (m Module) Dispatch(request app.Request, response app.Response) error {
	if len(request.Args) > 1 && request.Args[0] == "gif" {
		tag := strings.Join(request.Args[1:], " ")
		return m.Gif(request, response, tag)
	}

	return nil
}

func (m Module) Gif(request app.Request, response app.Response, tag string) error {
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

	embed := discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: gif.URL,
		},
	}

	_, err = response.Session.ChannelMessageSendEmbed(response.ChannelId, &embed)
	if err != nil {
		return fmt.Errorf("unable to send message channel: %w", err)
	}

	return nil
}
