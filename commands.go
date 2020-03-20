package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MalSearchResponse struct {
	Results []struct {
		Type string `json:"type"`
		Title string `json:"title"`
		ImageURL string `json:"image_url"`
		PageURL string `json:"url"`
		Description string `json:"synopsis"`
		Score float64 `json:"score"`
		EpisodeCount int `json:"episodes"`
		Airing bool `json:"airing"`
		StartDate time.Time `json:"start_date,omitempty"`
		EndDate time.Time `json:"end_date,omitempty"`
	} `json:"results"`
}

func commandHelp(s *discordgo.Session, channelId string) {
	usage := "Thy lost?\n" +
		"\n" +
		"**Commands**\n" +
		"\t• `!help`\tThis exact message.\n" +
		"\t• `!mal <action> <args>`\tInteract with the MyAnimeList website."

	sendChannelMessage(s, channelId, usage)
	return
}

func commandMal(s *discordgo.Session, channelId string, args []string) {
	if len(args) == 0 {
		goto usage
	}

	if args[0] == "search" && len(args) > 1 {
		name := strings.Join(args[1:], " ")
		commandMalSearch(s, channelId, name)
		return
	}

usage:
	usage := "Command `!mal` was malformed.\n" +
		"\n" +
		"**Usage**\n" +
		"`!map <action> <args>`\n" +
		"\n" +
		"**Actions**\n" +
		"`!mal search <name>`\tSearch an anime by name in the database."

	sendChannelMessage(s, channelId, usage)
	return
}

func commandMalSearch(s *discordgo.Session, channelId string, name string) {
	response, err := http.Get("https://api.jikan.moe/v3/search/anime?q=" + url.QueryEscape(name) + "&limit=1")
	if err != nil {
		sendChannelMessage(s, channelId, "Unable to contact MAL's API: " + err.Error())
		return
	}

	searchResponse := MalSearchResponse{}

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&searchResponse)
	if err != nil {
		sendChannelMessage(s, channelId, "Invalid JSON response from MAL's API: " + err.Error())
		return
	}

	if len(searchResponse.Results) == 0 {
		sendChannelMessage(s, channelId, "No results found")
		return
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
					Value: formatDate(searchResponse.Results[0].StartDate),
					Inline: true,
				},
				{
					Name: "End date",
					Value: formatDate(searchResponse.Results[0].EndDate),
					Inline: true,
				},
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(channelId, &message)
	if err != nil {
		log.Println("Unable to send channel embed message:", err)
	}
}
