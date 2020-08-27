package deejay

import (
	"fmt"
	"senko/app"
	"senko/modules/youtube"
	"strings"
)

// TODO: Queueing mechanism.
type Deejay struct {
	stopper chan struct{}
}

func (dj *Deejay) OnRegister(store *app.Store) {}

func (dj *Deejay) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if vars, ok := e.Match("play <what>"); ok {
			what := vars["what"]

			if !strings.HasPrefix(what, "http") {
				what = fmt.Sprintf("ytsearch1:%s", what)
			}

			mp3Filepath, err := youtube.DownloadAsMp3(what)
			if err != nil {
				return err
			}

			normalizedMp3, err := youtube.NormalizeForLoudness(mp3Filepath)
			if err != nil {
				return err
			}

			dj.stopper, err = gateway.PlayAudioFile(e.GuildID, normalizedMp3)
			if err != nil {
				return err
			}

			return nil
		}

		if _, ok := e.Match("stop"); ok {
			if dj.stopper != nil {
				close(dj.stopper)
			}
		}
	}

	return nil
}
