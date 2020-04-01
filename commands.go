package main

import (
	"bufio"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/url"
	"os/exec"
	"senko/libmal"
	"senko/libtenor"
	"strings"
	"time"
)

func commandHelp(s *discordgo.Session, channelId string) error {
	_, err := s.ChannelMessageSend(channelId, "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

func commandAnime(s *discordgo.Session, channelId string, args []string) error {
	if len(args) > 1 && args[0] == "search" {
		name := strings.Join(args[1:], " ")
		return commandAnimeSearch(s, channelId, name)
	}

	return nil
}

func commandAnimeSearch(s *discordgo.Session, channelId string, name string) error {
	mal := libmal.NewMal()
	searchResponse, err := mal.SearchAnime(name)
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
	tenorToken := loadConfig("TENOR_TOKEN")
	tenor := libtenor.NewTenor(tenorToken)
	tag := strings.Join(args, " ")

	channel, err := s.Channel(channelId)
	if err != nil {
		return fmt.Errorf("unable to lookup channel by id: %w", err)
	}

	if args[0] == "-nsfw" || channel.NSFW {
		tenor.NSFW = true
		tag = strings.Join(args[1:], " ")
	}

	gif, err := tenor.RandomGif(tag)
	if err != nil {
		return fmt.Errorf("unable to contact tenor: %w", err)
	}

	embed := discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: gif.URL,
		},
	}

	_, err = s.ChannelMessageSendEmbed(channelId, &embed)
	if err != nil {
		return fmt.Errorf("unable to send message channel: %w", err)
	}

	return nil
}

func commandYoutube(s *discordgo.Session, channelId string, args []string) error {
	if len(args) == 2 && args[0] == "download" {
		return commandYoutubeDownload(s, channelId, args[1])
	}

	return nil
}

func commandYoutubeDownload(s *discordgo.Session, channelId string, youtubeUrl string) error {
	msg, err := s.ChannelMessageSend(channelId, "Initializing download...")
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	args := []string {
		"-f",
		"bestvideo+bestaudio",
		"--newline",
		youtubeUrl,
	}

	cmd := exec.Command("youtube-dl", args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to create pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start youtube-dl: %w", err)
	}

	buffer := bufio.NewReader(out)

	filename := ""
	previousPercentage := 0.0

	for {
		rawLine, _, err := buffer.ReadLine()
		if err != nil {
			break
		}

		line := string(rawLine)
		size := ""
		percentage := 0.0

		_, err = fmt.Sscanf(line, "[download] %f%% of %s", &percentage, &size)
		if err == nil {
			if percentage - previousPercentage > 10 || (percentage == 100 && previousPercentage != 100) {
				content := fmt.Sprintf("Downloading video (%0.2f%% of %s)...", percentage, size)
				msg, err = s.ChannelMessageEdit(channelId, msg.ID, content)
				if err != nil {
					return fmt.Errorf("unable to send channel message: %w", err)
				}

				previousPercentage = percentage
			}
		}

		_, _ = fmt.Sscanf(line, "[ffmpeg] Merging formats into %q", &filename)

		if strings.HasSuffix(line, "has already been downloaded and merged") {
			filename = line
			filename = strings.TrimPrefix(filename, "[download] ")
			filename = strings.TrimSuffix(filename, " has already been downloaded and merged")
		}
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("unable to wait for youtube-dl: %w", err)
	}

	err = s.ChannelMessageDelete(channelId, msg.ID)
	if err != nil {
		return fmt.Errorf("unable to delete channel message: %w", err)
	}

	finalUrl := loadConfig("EXTERNAL_URL_PREFIX") + "/downloads/" + url.QueryEscape(filename)

	message := discordgo.MessageSend{
		Content: fmt.Sprintf("Downloaded `%s`!", filename),
		Embed: &discordgo.MessageEmbed{
			Timestamp:   time.Now().Format(time.RFC3339),
			Title:       "Direct link",
			URL:         finalUrl,
		},
	}

	_, err = s.ChannelMessageSendComplex(channelId, &message)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}
