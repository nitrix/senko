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
// TODO: Periodically update youtube-dl and ffmpeg while running?
// TODO: Merge deejay and youtube?

type Youtube struct{}

func (y *Youtube) OnRegister(store *app.Store) {}

func (y *Youtube) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if vars, ok := e.Match("youtube download <target>"); ok {
			return y.download(gateway, e.ChannelID, vars["target"])
		}

		if vars, ok := e.Match("youtube mp3 <target>"); ok {
			return y.mp3(gateway, e.ChannelID, vars["target"])
		}
	}

	return nil
}

func (y Youtube) download(gateway *app.Gateway, channelID app.ChannelID, youtubeUrl string) error {
	err := gateway.SendMessage(channelID, "Downloading...")
	if err != nil {
		return err
	}

	args := []string{
		"-f",
		"bestvideo+bestaudio",
		"--write-info-json",
		"--newline",
		"-o", "downloads/%(title)s-%(id)s-%(epoch)s.%(ext)s",
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

		if strings.HasPrefix(line, "[download] Destination:") {
			mediaFilepath = strings.TrimPrefix(line, "[download] Destination: ")
		}

		if strings.HasPrefix(line, "[ffmpeg] Merging formats into ") {
			mediaFilepath = strings.TrimPrefix(line, "[ffmpeg] Merging formats into ")
			mediaFilepath = strings.TrimLeft(mediaFilepath, "\"")
			mediaFilepath = strings.TrimRight(mediaFilepath, "\"")
		}
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("unable to wait for youtube-dl: %w", err)
	}

	metadataFilepath := strings.TrimSuffix(mediaFilepath, filepath.Ext(mediaFilepath)) + ".info.json"

	metadataFilepath = filepath.ToSlash(metadataFilepath)
	mediaFilepath = filepath.ToSlash(mediaFilepath)

	mediaLink := gateway.GetEnv("EXTERNAL_URL_PREFIX") + "/" + filepath.ToSlash(filepath.Dir(mediaFilepath)) + "/" + url.PathEscape(filepath.Base(mediaFilepath))
	metadataLink := gateway.GetEnv("EXTERNAL_URL_PREFIX") + "/" + filepath.ToSlash(filepath.Dir(metadataFilepath)) + "/" + url.PathEscape(filepath.Base(metadataFilepath))

	metadataFile, err := os.Open(metadataFilepath)
	if err != nil {
		return fmt.Errorf("unable to open metadata file: %w", err)
	}

	defer func() {
		_ = metadataFile.Close()
	}()

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

	return gateway.SendEmbed(channelID, discordgo.MessageEmbed{
		Title: title,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Media link", Value: mediaLink, Inline: true},
			{Name: "Metadata link", Value: metadataLink, Inline: true},
		},
	})
}

func (y Youtube) mp3(gateway *app.Gateway, channelID app.ChannelID, youtubeUrl string) error {
	err := gateway.SendMessage(channelID, "Downloading...")
	if err != nil {
		return err
	}

	filePath, err := DownloadAsMp3(youtubeUrl)
	if err != nil {
		return err
	}

	return gateway.SendFile(channelID, filePath)
}

func DownloadAsMp3(youtubeUrl string) (string, error) {
	args := []string{
		"-f", "bestaudio",
		"--extract-audio",
		"--audio-format", "mp3",
		"--newline",
		"-o", "downloads/%(title)s-%(id)s-%(epoch)s.%(ext)s",
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

func NormalizeForLoudness(filepath string) (string, error) {
	normalizedFilepath := filepath + ".norm.mp3"

	args := []string{
		"-i",
		filepath,
		"-filter:a",
		"loudnorm",
		normalizedFilepath,
	}

	cmd := exec.Command("ffmpeg", args...)

	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("unable to start ffmpeg loudnorm: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return "", fmt.Errorf("unable to wait for ffmpeg: %w", err)
	}

	return normalizedFilepath, nil
}
