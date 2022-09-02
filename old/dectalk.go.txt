package dectalk

import (
	"bytes"
	"net/http"
	"os/exec"
	"senko/app"
)

type Dectalk struct{}

func (d *Dectalk) OnRegister(store *app.Store) {}

func (d *Dectalk) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if vars, ok := e.Match("dectalk <text>"); ok {
			return d.dectalk(gateway, e.GuildID, vars["text"])
		}
	}

	return nil
}

func (d *Dectalk) dectalk(gateway *app.Gateway, guildID app.GuildID, text string) error {
	body := bytes.NewBufferString(text)
	response, err := http.Post("https://dectalk.nitrix.me/synthesize", "plain/text", body)
	if err != nil {
		return err
	}

	ffmpegArgs := []string{
		"-i",
		"-",
		"-f",
		"s16le",
		"-acodec",
		"pcm_s16le",
		"-ac",
		"2",
		"-ar",
		"48000",
		"-filter:a",
		"loudnorm",
		"pipe:1",
	}

	ffmpegCmd := exec.Command("ffmpeg", ffmpegArgs...)
	ffmpegCmd.Stdin = response.Body

	ffmpegPipe, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = ffmpegCmd.Start()
	if err != nil {
		return err
	}

	_, err = gateway.PlayAudioStream(guildID, ffmpegPipe)
	if err != nil {
		return err
	}

	return nil
}