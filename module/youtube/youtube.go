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

// TODO: This isn't actually limited to youtube. Youtube-dl supports many more sites.

type Youtube struct{}

func (y *Youtube) OnLoad(store *app.Store) {}

func (y *Youtube) OnUnload(store *app.Store) {}

func (y *Youtube) OnEvent(event *app.Event) error {
	if event.Kind == app.CommandEvent {
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
	}

	return nil
}

func (y Youtube) download(event *app.Event, youtubeUrl string) error {
	args := []string{
		"-f",
		"bestvideo+bestaudio",
		"--write-info-json",
		"--newline",
		"-o", "downloads/%(title)s-%(id)s.%(ext)s",
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

	mediaFilepath := ""

	for {
		rawLine, _, err := buffer.ReadLine()
		if err != nil {
			break
		}

		line := string(rawLine)

		_, _ = fmt.Sscanf(line, "[ffmpeg] Merging formats into %q", &mediaFilepath)

		if strings.HasSuffix(line, "has already been downloaded and merged") {
			mediaFilepath = line
			mediaFilepath = strings.TrimPrefix(mediaFilepath, "[download] ")
			mediaFilepath = strings.TrimSuffix(mediaFilepath, " has already been downloaded and merged")
		}
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("unable to wait for youtube-dl: %w", err)
	}

	metadataFilepath := strings.TrimSuffix(mediaFilepath, filepath.Ext(mediaFilepath)) + ".info.json"

	mediaLink := app.GetEnvironmentVariable("EXTERNAL_URL_PREFIX") + "/" + filepath.ToSlash(filepath.Dir(mediaFilepath)) + "/" + url.PathEscape(filepath.Base(mediaFilepath))
	metadataLink := app.GetEnvironmentVariable("EXTERNAL_URL_PREFIX") + "/" + filepath.ToSlash(filepath.Dir(metadataFilepath)) + "/" + url.PathEscape(filepath.Base(metadataFilepath))

	metadataFile, err := os.Open(metadataFilepath)
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

func (y Youtube) mp3(event *app.Event, youtubeUrl string) error {
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

	err = normalizeForLoudness(mp3Filename)
	if err != nil {
		return "", err
	}

	return mp3Filename, nil
}

func normalizeForLoudness(filepath string) error {
	err := os.Rename(filepath, filepath + ".bak")
	if err != nil {
		return err
	}

	args := []string{
		"-i",
		filepath + ".bak",
		"-filter:a",
		"loudnorm",
		filepath,
	}

	cmd := exec.Command("ffmpeg", args...)

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start ffmpeg loudnorm: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("unable to wait for ffmpeg: %w", err)
	}

	return nil
}
