package youtube

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"senko/app"
	"strings"
)

type Module struct {}

func (m Module) Dispatch(request app.Request, response app.Response) error {
	if len(request.Args) == 3 && request.Args[0] == "youtube" && request.Args[1] == "download" {
		return m.download(response, request.Args[2])
	}

	if len(request.Args) == 3 && request.Args[0] == "youtube" && request.Args[1] == "mp3" {
		return m.mp3(response, request.Args[2])
	}

	return nil
}

func (m Module) download(response app.Response, youtubeUrl string) error {
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

	return response.SendEmbed(app.Embed{
		Title: title,
		Fields: []app.Field {
			{ Key: "Media link", Value: mediaLink },
			{ Key: "Metadata link", Value: metadataLink },
		},
	})
}

func (m Module) mp3(response app.Response, youtubeUrl string) error {
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

	return response.SendFile("downloads/" + mp3Filename)
}
