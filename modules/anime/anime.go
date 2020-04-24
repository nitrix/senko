package anime

import (
	"fmt"
	"senko/app"
	"senko/modules/anime/mal"
	"strings"
)

type Module struct {}

func (m Module) Dispatch(req app.Request, resp app.Response) error {
	if len(req.Args) > 2 && req.Args[0] == "anime" && req.Args[1] == "search" {
		name := strings.Join(req.Args[2:], " ")
		return m.search(resp, name)
	}

	return nil
}

func (m Module) search(resp app.Response, name string) error {
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

	return resp.SendEmbed(app.Embed{
		Message: "Closest match found on MyAnimeList.",
		Title:       searchResponse.Results[0].Title,
		URL:         searchResponse.Results[0].PageURL,
		Description: searchResponse.Results[0].Description,
		ThumbnailURL: searchResponse.Results[0].ImageURL,
		Fields: []app.Field {
			{ Key: "Type", Value: searchResponse.Results[0].Type },
			{ Key: "Episodes", Value: fmt.Sprint(searchResponse.Results[0].EpisodeCount) },
			{ Key: "Score", Value: fmt.Sprint(searchResponse.Results[0].Score) },
			{ Key: "Airing", Value: airing },
			{ Key: "Start date", Value: app.FormatDate(searchResponse.Results[0].StartDate) },
			{ Key: "End date", Value: app.FormatDate(searchResponse.Results[0].EndDate) },
		},
	})
}
