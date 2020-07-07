package youtube

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
	"senko/app"
	"strings"
)

type Youtube struct{}

func (y *Youtube) Load() error { return nil }

func (y *Youtube) Unload() error { return nil }

func (y *Youtube) OnCommand(event *app.CommandEvent) error {
	if !strings.HasPrefix(event.Content, "youtube ") {
		return nil
	}

	parts := strings.Split(strings.TrimPrefix(event.Content, "youtube "), " ")

	if len(parts) == 2 && parts[0] == "download" {
		return y.download(event, parts[1])
	}

	if len(parts) == 2 && parts[0] == "mp3" {
		return y.mp3(event, parts[1])
	}

	return nil
}

func (y *Youtube) OnMessageCreated(event *app.MessageCreatedEvent) error { return nil }

func (y Youtube) download(event *app.CommandEvent, youtubeUrl string) error {
	args := []string{
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
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("unable to wait for youtube-dl: %w", err)
	}

	metadataFilename := strings.TrimSuffix(mediaFilename, filepath.Ext(mediaFilename)) + ".info.json"
	mediaLink := app.GetToken("EXTERNAL_URL_PREFIX") + "/downloads/" + url.QueryEscape(mediaFilename)
	metadataLink := app.GetToken("EXTERNAL_URL_PREFIX") + "/downloads/" + url.QueryEscape(metadataFilename)

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

	return event.ReplyEmbed(discordgo.MessageEmbed{
		Title: title,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Media link", Value: mediaLink, Inline: true},
			{Name: "Metadata link", Value: metadataLink, Inline: true},
		},
	})
}

func (y Youtube) mp3(event *app.CommandEvent, youtubeUrl string) error {
	filePath, err := DownloadAsMp3(youtubeUrl)
	if err != nil {
		return err
	}

	return event.ReplyFile(filePath)
}

func DownloadAsMp3(youtubeUrl string) (string, error) {
	args := []string{
		"-f", "bestaudio",
		"--extract-audio",
		"--audio-format", "mp3",
		"--newline",
		"-o", "downloads/%(title)s-%(id)s.%(ext)s",
		youtubeUrl,
	}

	cmd := exec.Command("youtube-dl", args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("unable to create pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return "", fmt.Errorf("unable to start youtube-dl: %w", err)
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
		return "", fmt.Errorf("unable to wait for youtube-dl: %w", err)
	}

	return mp3Filename, nil
}
