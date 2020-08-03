package anime

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"senko/app"
	"senko/module/anime/mal"
	"strings"
)

type Anime struct{}

func (a *Anime) OnLoad(store *app.Store) {}

func (a *Anime) OnUnload(store *app.Store) {}

func (a *Anime) OnEvent(event *app.Event) error {
	if event.Kind == app.CommandEvent {
		if !strings.HasPrefix(event.Content, "anime ") {
			return nil
		}

		command := strings.TrimPrefix(event.Content, "anime ")
		parts := strings.Split(command, " ")

		if len(parts) > 1 && parts[0] == "search" {
			name := strings.Join(parts[1:], " ")
			return a.search(event, name)
		}
	}

	return nil
}

func (a *Anime) search(event *app.Event, name string) error {
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

	return event.ReplyComplex(discordgo.MessageSend{
		Content: "Closest match found on MyAnimeList.",
		Embed: &discordgo.MessageEmbed{
			Title:       searchResponse.Results[0].Title,
			URL:         searchResponse.Results[0].PageURL,
			Description: searchResponse.Results[0].Description,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: searchResponse.Results[0].ImageURL,
			},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Type", Value: searchResponse.Results[0].Type, Inline: true},
				{Name: "Episodes", Value: fmt.Sprint(searchResponse.Results[0].EpisodeCount), Inline: true},
				{Name: "Score", Value: fmt.Sprint(searchResponse.Results[0].Score), Inline: true},
				{Name: "Airing", Value: airing, Inline: true},
				{Name: "Start date", Value: app.FormatDate(searchResponse.Results[0].StartDate), Inline: true},
				{Name: "End date", Value: app.FormatDate(searchResponse.Results[0].EndDate), Inline: true},
			},
		},
	})
}
