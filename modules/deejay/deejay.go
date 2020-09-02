package deejay

import (
	"fmt"
	"log"
	"os/exec"
	"senko/app"
	"strings"
	"sync"
)

type Deejay struct {
	mutex sync.Mutex
	stoppers map[app.GuildID]chan struct{}
	//queue chan string
}

func (dj *Deejay) OnRegister(store *app.Store) {
	dj.stoppers = make(map[app.GuildID]chan struct{})
	//dj.queue = make(chan string, 0)
}

func (dj *Deejay) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if vars, ok := e.Match("play <what>"); ok {
			return dj.play(gateway, e.GuildID, vars["what"])
		}

		if _, ok := e.Match("stop"); ok {
			dj.stop(e.GuildID)
		}

		if vars, ok := e.Match("queue <what>"); ok {
			dj.addToQueue(vars["what"])
		}
	}

	return nil
}

func (dj *Deejay) addToQueue(what string) {
	//dj.queue <- what
}

func (dj *Deejay) stop(guildId app.GuildID) {
	dj.mutex.Lock()
	defer dj.mutex.Unlock()

	if dj.stoppers[guildId] != nil {
		close(dj.stoppers[guildId])
		delete(dj.stoppers, guildId)
	}
}

func (dj *Deejay) play(gateway *app.Gateway, guildID app.GuildID, what string) error {
	// Stop the previously playing thing.
	dj.stop(guildID)

	if !strings.HasPrefix(what, "http") {
		what = fmt.Sprintf("ytsearch1:%s", what)
	}

	youtubeArgs := []string{
		"-f",
		"bestaudio",
		"--audio-format",
		"wav",
		what,
		"-q",
		"-o",
		"-",
	}

	youtubeCmd := exec.Command("youtube-dl", youtubeArgs...)

	youtubePipe, err := youtubeCmd.StdoutPipe()
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
	ffmpegCmd.Stdin = youtubePipe

	ffmpegPipe, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = youtubeCmd.Start()
	if err != nil {
		return err
	}

	err = ffmpegCmd.Start()
	if err != nil {
		return err
	}

	stopper, err := gateway.PlayAudioStream(guildID, ffmpegPipe)
	if err != nil {
		log.Fatalln(err)
	}

	dj.mutex.Lock()
	dj.stoppers[guildID] = stopper
	dj.mutex.Unlock()

	go func () {
		<-stopper

		_ = youtubeCmd.Process.Kill()
		_ = ffmpegCmd.Process.Kill()
	}()

	return nil
}