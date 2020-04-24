package anime

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"senko/app"
	"senko/modules/anime/mal"
	"strings"
)

type Module struct {}

func (m Module) Dispatch(req app.Request, resp app.Response) error {
	if len(req.Args) > 2 && req.Args[0] == "anime" && req.Args[1] == "search" {
		name := strings.Join(req.Args[2:], " ")
		return m.Search(resp, name)
	}

	return nil
}

func (m Module) Search(resp app.Response, name string) error {
	malInstance := mal.NewMal()
	searchResponse, err := malInstance.SearchAnime(name)
	if err != nil {
		return fmt.Errorf("unable to search anime on MAL: %w", err)
	}

	if len(searchResponse.Results) == 0 {
		return fmt.Errorf("no results found")
	}

	airing := "No"
	if searchResponse.Results[0].Airing {
		airing = "Yes"
	}

	message := discordgo.MessageSend{
		Content: "Closest match found on MyAnimeList.",
		Embed: &discordgo.MessageEmbed{
			Title:       searchResponse.Results[0].Title,
			URL:         searchResponse.Results[0].PageURL,
			Description: searchResponse.Results[0].Description,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: searchResponse.Results[0].ImageURL,
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Type",
					Value: searchResponse.Results[0].Type,
					Inline: true,
				},
				{
					Name: "Episodes",
					Value: fmt.Sprint(searchResponse.Results[0].EpisodeCount),
					Inline: true,
				},
				{
					Name: "Score",
					Value: fmt.Sprint(searchResponse.Results[0].Score),
					Inline: true,
				},
				{
					Name: "Airing",
					Value: airing,
					Inline: true,
				},
				{
					Name: "Start date",
					Value: app.FormatDate(searchResponse.Results[0].StartDate),
					Inline: true,
				},
				{
					Name: "End date",
					Value: app.FormatDate(searchResponse.Results[0].EndDate),
					Inline: true,
				},
			},
		},
	}

	_, err = resp.Session.ChannelMessageSendComplex(resp.ChannelId, &message)
	if err != nil {
		return fmt.Errorf("unable to send channel embed message: %w", err)
	}

	return nil
}
