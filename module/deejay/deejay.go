package deejay

import (
	"fmt"
	"senko/app"
	"senko/module/youtube"
	"strings"
)

// TODO: Queueing mechanism.
type Deejay struct {
	stopper chan struct{}
}

func (dj *Deejay) OnLoad(store *app.Store) {}

func (dj *Deejay) OnUnload(store *app.Store) {}

func (dj *Deejay) OnEvent(event *app.Event) error {
	if event.Kind == app.CommandEvent {
		if strings.HasPrefix(event.Content, "play ") {
			what := strings.TrimPrefix(event.Content, "play ")

			if !strings.HasPrefix(what, "http") {
				what = fmt.Sprintf("ytsearch1:%s", what)
			}

			mp3Filepath, err := youtube.DownloadAsMp3(what)
			if err != nil {
				return err
			}

			dj.stopper, err = event.PlayAudioFile(mp3Filepath)
			if err != nil {
				return err
			}

			return nil
		}

		if event.Content == "stop" {
			if dj.stopper != nil {
				close(dj.stopper)
			}
		}
	}

	return nil
}
