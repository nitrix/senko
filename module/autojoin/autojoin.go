package autojoin

import (
	"fmt"
	"senko/app"
	"strings"
	"sync"
)

type Autojoin struct {
	mutex sync.RWMutex
	channelId string
}

func (a *Autojoin) OnLoad(store *app.Store) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	channelId := store.Read("autojoin.channelId")

	switch v := channelId.(type) {
	case string:
		a.channelId = v
	default:
		a.channelId = ""
	}
}

func (a *Autojoin) OnUnload(store *app.Store) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	store.Write("autojoin.channelId", a.channelId)
}

func (a *Autojoin) OnEvent(event *app.Event) error {
	if event.Kind == app.CommandEvent {
		if strings.HasPrefix(event.Content, "autojoin enable ") {
			channelName := strings.TrimPrefix(event.Content, "autojoin enable ")

			channel, err := event.FindChannelByName(channelName)
			if err != nil {
				return err
			}

			a.mutex.Lock()
			a.channelId = channel.ID
			a.mutex.Unlock()

			err = event.Reply("Autojoin enabled for channel " + channel.Name)
			if err != nil {
				return err
			}

			err = a.autojoin(event)
			if err != nil {
				return err
			}
		}

		if event.Content == "autojoin disable" {
			a.mutex.Lock()
			a.channelId = ""
			a.mutex.Unlock()
		}
	}

	switch event.Kind {
	case app.VoiceJoinEvent: fallthrough
	case app.VoiceLeaveEvent: fallthrough
	case app.CurrentlyInVoiceEvent: fallthrough
	case app.ReadyEvent:
		fmt.Println("-->", event.Kind)
		return a.autojoin(event)
	}

	return nil
}

func (a *Autojoin) autojoin(event *app.Event) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	inUse, err := event.IsChannelInUse(a.channelId)
	if err != nil {
		return err
	}

	if inUse {
		return event.JoinVoice(a.channelId)
	} else {
		return event.LeaveVoice(a.channelId)
	}
}