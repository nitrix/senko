package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sanzaru/go-giphy"
	"io/ioutil"
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

func commandHelp(s *discordgo.Session, channelId string) error {
	_, err := s.ChannelMessageSend(channelId, "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

func commandMal(s *discordgo.Session, channelId string, args []string) error {
	if len(args) > 1 && args[0] == "search" {
		name := strings.Join(args[1:], " ")
		return commandMalSearch(s, channelId, name)
	} else if len(args) > 2 && args[0] == "cross" {
		// return commandMalCross(s, channelId, args[1], args[2])
	}

	return nil
}

func commandMalCross(s *discordgo.Session, channelId string, source string, target string) error {
	response, err := http.Get("https://api.jikan.moe/v3/user/nitrixen/animelist/completed")
	if err != nil {
		return fmt.Errorf("unable to contact MAL's API: %w", err)
	}

	content, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(content))

	return nil
}

func commandMalSearch(s *discordgo.Session, channelId string, name string) error {
	response, err := http.Get("https://api.jikan.moe/v3/search/anime?q=" + url.QueryEscape(name) + "&limit=1")
	if err != nil {
		return fmt.Errorf("unable to contact MAL's API: %w", err)
	}

	searchResponse := MalSearchResponse{}

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&searchResponse)
	if err != nil {
		return fmt.Errorf("invalid JSON response from MAL's API: %w", err)
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
		return fmt.Errorf("unable to send channel embed message: %w", err)
	}

	return nil
}

func commandGif(s *discordgo.Session, channelId string, args []string) error {
	giphyToken := loadToken("GIPHY_TOKEN")
	giphy := libgiphy.NewGiphy(giphyToken)
	tag := strings.Join(args, " ")

	result, err := giphy.GetRandom(tag)
	if err != nil {
		return fmt.Errorf("unable to contact Giphy: %w", err)
	}

	embed := discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: result.Data.Image_original_url,
		},
	}

	_, err = s.ChannelMessageSendEmbed(channelId, &embed)
	if err != nil {
		return fmt.Errorf("unable to contact Giphy: %w", err)
	}

	return nil
}