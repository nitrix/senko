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

type Plugin struct{}

func (p *Plugin) Save() error    { return nil }
func (p *Plugin) Restore() error { return nil }

func (p Plugin) OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) error {
	if !strings.HasPrefix(message.Content, "!youtube ") {
		return nil
	}

	parts := strings.Split(strings.TrimPrefix(message.Content, "!youtube "), " ")

	if len(parts) == 2 && parts[0] == "download" {
		return p.download(session, message.ChannelID, parts[1])
	}

	if len(parts) == 2 && parts[0] == "mp3" {
		return p.mp3(session, message.ChannelID, parts[1])
	}

	return nil
}

func (p Plugin) download(session *discordgo.Session, channelId string, youtubeUrl string) error {
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

	_, err = session.ChannelMessageSendEmbed(channelId, &discordgo.MessageEmbed{
		Title: title,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Media link", Value: mediaLink, Inline: true},
			{Name: "Metadata link", Value: metadataLink, Inline: true},
		},
	})

	return err
}

func (p Plugin) mp3(session *discordgo.Session, channelId string, youtubeUrl string) error {
	filePath, err := DownloadAsMp3(youtubeUrl)
	if err != nil {
		return err
	}

	return app.DiscordSendFile(session, channelId, filePath)
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
