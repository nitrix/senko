package autojoin

import (
	"senko/app"
	"strings"
	"sync"
)

type Autojoin struct {
	mutex sync.RWMutex
	channelId string
}

func (a *Autojoin) OnRegister(store *app.Store) {
	store.Link("autojoin.channelId", &a.channelId, "")
}

func (a *Autojoin) OnEvent(gateway app.Gateway, event interface{}) error {
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