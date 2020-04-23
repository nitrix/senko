package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"senko/pkg/mal"
	"senko/pkg/tenor"
	"strings"
)

func commandVersion(s *discordgo.Session, channelId string) error {
	_, err := s.ChannelMessageSend(channelId, Version)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

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
	tenorToken := getToken("TENOR_TOKEN")
	tenorInstance := tenor.NewTenor(tenorToken)
	tag := strings.Join(args, " ")

	channel, err := s.Channel(channelId)
	if err != nil {
		return fmt.Errorf("unable to lookup channel by id: %w", err)
	}

	if args[0] == "-nsfw" || channel.NSFW {
		tenorInstance.NSFW = true
		tag = strings.Join(args[1:], " ")
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

	if len(args) == 2 && args[0] == "mp3" {
		return commandYoutubeMp3(s, channelId, args[1])
	}

	return nil
}

func commandYoutubeDownload(s *discordgo.Session, channelId string, youtubeUrl string) error {
	msg, err := s.ChannelMessageSend(channelId, "Downloading video...")
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	_ = os.Mkdir("downloads", 0644)

	args := []string {
		"-f",
		"bestvideo+bestaudio",
		"--write-info-json",
		"--newline",
		youtubeUrl,
	}

	cmd := exec.Command("youtube-dl", args...)
	cmd.Dir = "downloads"
	out, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to create pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start youtube-dl: %w", err)
	}

	buffer := bufio.NewReader(out)

	mediaFilename := ""

	for {
		rawLine, _, err := buffer.ReadLine()
		if err != nil {
			break
		}

		line := string(rawLine)

		_, _ = fmt.Sscanf(line, "[ffmpeg] Merging formats into %q", &mediaFilename)

		if strings.HasSuffix(line, "has already been downloaded and merged") {
			mediaFilename = line
			mediaFilename = strings.TrimPrefix(mediaFilename, "[download] ")
			mediaFilename = strings.TrimSuffix(mediaFilename, " has already been downloaded and merged")
		}

		// _, _ = s.ChannelMessageEdit(channelId, msg.ID, fmt.Sprintf("Downloading...\n```%s```", strings.Join(lines, "\n")))
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("unable to wait for youtube-dl: %w", err)
	}

	err = s.ChannelMessageDelete(channelId, msg.ID)
	if err != nil {
		return fmt.Errorf("unable to delete channel message: %w", err)
	}

	metadataFilename := strings.TrimSuffix(mediaFilename, filepath.Ext(mediaFilename)) + ".info.json"
	mediaLink := getToken("EXTERNAL_URL_PREFIX") + "/downloads/" + url.QueryEscape(mediaFilename)
	metadataLink := getToken("EXTERNAL_URL_PREFIX") + "/downloads/" + url.QueryEscape(metadataFilename)

	metadataFile, err := os.Open("downloads/" + metadataFilename)
	if err != nil {
		return fmt.Errorf("unable to open metadata file: %w", err)
	}
	content, err := ioutil.ReadAll(metadataFile)
	if err != nil {
		return fmt.Errorf("unable to real metadata file: %w", err)
	}
	metadata := make(map[string]interface{})
	err = json.Unmarshal(content, &metadata)
	if err != nil {
		return fmt.Errorf("unable to unmarshal metadata: %w", err)
	}

	title, ok := metadata["title"].(string)
	if !ok {
		return errors.New("title must be a string")
	}

	message := discordgo.MessageSend {
		Content: "Download complete!",
		Embed: &discordgo.MessageEmbed{
			Title:     title,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Media link",
					Value: mediaLink,
				},
				{
					Name: "Metadata link",
					Value: metadataLink,
				},
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(channelId, &message)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

func commandYoutubeMp3(s *discordgo.Session, channelId string, youtubeUrl string) error {
	msg, err := s.ChannelMessageSend(channelId, "Downloading mp3...")
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	_ = os.Mkdir("downloads", 0644)

	args := []string {
		"-f",
		"bestaudio",
		"--extract-audio",
		"--audio-format",
		"mp3",
		"--newline",
		youtubeUrl,
	}

	cmd := exec.Command("youtube-dl", args...)
	cmd.Dir = "downloads"
	out, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to create pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start youtube-dl: %w", err)
	}

	buffer := bufio.NewReader(out)

	mp3Filename := ""

	for {
		rawLine, _, err := buffer.ReadLine()
		if err != nil {
			break
		}

		line := string(rawLine)

		if strings.HasPrefix(line, "[ffmpeg] Destination:") {
			mp3Filename = strings.TrimPrefix(line, "[ffmpeg] Destination: ")
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

	mp3file, err := os.Open("downloads/" + mp3Filename)
	defer func() {
		_ = mp3file.Close()
	}()

	message := discordgo.MessageSend {
		Content: "Download complete!",
		Files: []*discordgo.File {
			{
				Name:        mp3Filename,
				ContentType: "audio/mp3",
				Reader:      mp3file,
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(channelId, &message)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}