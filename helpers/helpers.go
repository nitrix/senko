package helpers

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

var mixers map[string]*Mixer = map[string]*Mixer{}
var mutex sync.RWMutex

func FindChannelByName(s *discordgo.Session, guildID string, channelName string) (*discordgo.Channel, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get channels: %w", err)
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel, nil
		}
	}

	return nil, errors.New("unable to find channel")
}

// TODO: Surely something useful to do with the incoming audio?
// pcmIncoming := make(chan *discordgo.Packet)
// go dgvoice.ReceivePCM(voiceConnection, pcmIncoming)

func VoiceConnectionToMixer(voiceConnection *discordgo.VoiceConnection) *Mixer {
	mutex.Lock()
	defer mutex.Unlock()

	mixer, ok := mixers[voiceConnection.GuildID]

	if !ok {
		mixer = NewMixer()

		go func() {
			dgvoice.SendPCM(voiceConnection, mixer.Output())
		}()

		mixers[voiceConnection.GuildID] = mixer
	}

	return mixer
}
