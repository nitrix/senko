package eggplant

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type Plugin struct {
	victims      []string
	victimsMutex sync.Mutex
}

func (p *Plugin) Save() error {
	p.victimsMutex.Lock()
	defer p.victimsMutex.Unlock()

	bytes, err := json.Marshal(p.victims)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("config/victims.txt", bytes, 0644)
}

func (p *Plugin) Restore() error {
	p.victimsMutex.Lock()
	defer p.victimsMutex.Unlock()

	file, err := os.Open("config/victims.txt")
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	return decoder.Decode(&p.victims)
}

func (p *Plugin) OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) error {
	p.check(session, message.ChannelID, message.ID, message.Content, message.Author.ID)

	if !strings.HasPrefix(message.Content, "!eggplant ") {
		return nil
	}

	parts := strings.Split(strings.TrimPrefix(message.Content, "!eggplant "), " ")

	if len(parts) == 2 && parts[0] == "enable" {
		return p.enable(session, message.ChannelID, message.GuildID, parts[1])
	}

	if len(parts) == 2 && parts[0] == "disable" {
		return p.disable(session, message.ChannelID, message.GuildID, parts[1])
	}

	if message.Content == "!help" {
		return p.help(session, message.ChannelID)
	}

	return nil
}

func (p *Plugin) check(session *discordgo.Session, channelId string, messageId string, messageContent string, userId string) {
	p.victimsMutex.Lock()
	defer p.victimsMutex.Unlock()

	if strings.Contains(strings.ToLower(messageContent), "o.o") {
		for _, victim := range p.victims {
			if victim == userId {
				p.eggplant(session, channelId, messageId)
			}
		}
	}
}

func (p *Plugin) eggplant(session *discordgo.Session, channelId string, messageId string) {
	_ = session.MessageReactionAdd(channelId, messageId, "🍆") // Eggplant
	_ = session.MessageReactionAdd(channelId, messageId, "🙄") // Rolling eyes
}

func (p *Plugin) enable(session *discordgo.Session, channelId, guildId string, nick string) error {
	p.victimsMutex.Lock()
	defer p.victimsMutex.Unlock()

	members, err := session.GuildMembers(guildId, "", 0)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.User.Username == nick || member.Nick == nick {
			// Make sure it's not already enabled
			for _, victim := range p.victims {
				if victim == member.User.ID {
					_, err = session.ChannelMessageSend(channelId, fmt.Sprintf("Eggplant already enabled for `%s`.", nick))
					return err
				}
			}

			p.victims = append(p.victims, member.User.ID)

			_, err = session.ChannelMessageSend(channelId, fmt.Sprintf("Eggplant enabled for `%s`.", nick))
			return err
		}
	}

	_, err = session.ChannelMessageSend(channelId, fmt.Sprintf("Nick `%s` not found.", nick))

	return nil
}

func (p *Plugin) disable(session *discordgo.Session, channelId, guildId string, nick string) error {
	p.victimsMutex.Lock()
	defer p.victimsMutex.Unlock()

	members, err := session.GuildMembers(guildId, "", 0)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.User.Username == nick || member.Nick == nick {
			// Make sure it's actually enabled.
			for i, victim := range p.victims {
				if victim == member.User.ID {
					p.victims[i] = p.victims[len(p.victims)-1]
					p.victims = p.victims[:len(p.victims)-1]

					_, err = session.ChannelMessageSend(channelId, fmt.Sprintf("Eggplant disabled for `%s`.", nick))
					return err
				}
			}

			_, err = session.ChannelMessageSend(channelId, fmt.Sprintf("Eggplant isn't enabled for `%s`.", nick))
			return err
		}
	}

	_, err = session.ChannelMessageSend(channelId, fmt.Sprintf("Nick `%s` not found.", nick))

	return nil
}

func (p Plugin) help(session *discordgo.Session, channelId string) error {
	msg := "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md"
	_, err := session.ChannelMessageSend(channelId, msg)
	return err
}
