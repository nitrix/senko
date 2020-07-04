package app

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"mime"
	"os"
	"path/filepath"
)

type MessageCreatedEvent struct {
	Content string
	AuthorId string

	session *discordgo.Session
	message *discordgo.MessageCreate
}

type CommandEvent struct {
	app *App
	Content string
	session *discordgo.Session
	channelId string
	guildId string
}

func (e MessageCreatedEvent) ReplyComplex(msg discordgo.MessageSend) error {
	_, err := e.session.ChannelMessageSendComplex(e.message.ChannelID, &msg)
	if err != nil {
		return err
	}

	return nil
}

func (e MessageCreatedEvent) Reply(msg string) error {
	_, err := e.session.ChannelMessageSend(e.message.ChannelID, msg)
	if err != nil {
		return err
	}

	return nil
}

func (e MessageCreatedEvent) React(emoji string) error {
	return e.session.MessageReactionAdd(e.message.ChannelID, e.message.ID, emoji)
}

func (e CommandEvent) ReplyComplex(msg discordgo.MessageSend) error {
	if e.channelId == "" {
		return errors.New("replying to a voice command unsupported") // FIXME
	}

	_, err := e.session.ChannelMessageSendComplex(e.channelId, &msg)
	if err != nil {
		return err
	}

	return nil
}

func (e CommandEvent) Reply(msg string) error {
	if e.channelId == "" {
		return errors.New("replying to a voice command unsupported") // FIXME
	}

	_, err := e.session.ChannelMessage(e.channelId, msg)
	if err != nil {
		return err
	}

	return nil
}

func (e CommandEvent) ResolveNick(nick string) (string, error) {
	members, err := e.session.GuildMembers(e.guildId, "", 0)
	if err != nil {
		return "", err
	}

	for _, member := range members {
		if member.User.Username == nick || member.Nick == nick {
			return member.User.ID, nil
		}
	}

	return "", errors.New("nick not found")
}

func (e CommandEvent) ReplyFile(path string) error {
	file, err := os.Open(path)
	defer func() {
		_ = file.Close()
	}()

	message := discordgo.MessageSend{
		Files: []*discordgo.File{
			{
				Name:        filepath.Base(path),
				ContentType: mime.TypeByExtension(filepath.Ext(path)),
				Reader:      file,
			},
		},
	}

	_, err = e.session.ChannelMessageSendComplex(e.channelId, &message)

	return err
}

func (e CommandEvent) ReplyEmbed(embed discordgo.MessageEmbed) error {
	_, err := e.session.ChannelMessageSendEmbed(e.channelId, &embed)
	if err != nil {
		return err
	}

	return nil
}

func (e CommandEvent) findChannelByName(name string) (*discordgo.Channel, error) {
	channels, err := e.session.GuildChannels(e.guildId)
	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if channel.Name == name {
			return channel, nil
		}
	}

	return nil, errors.New("channel not found")
}

func (e CommandEvent) DoCommand(command string) error {
	e.Content = command

	// FIXME: merge errors into final error

	for _, module := range e.app.modules {
		go module.OnCommand(&e)
	}

	return nil
}

func (e CommandEvent) JoinVoice(name string) (*discordgo.VoiceConnection, error) {
	channel, err := e.findChannelByName(name)
	if err != nil {
		return nil, err
	}

	return e.session.ChannelVoiceJoin(e.guildId, channel.ID, false, false)
}

func (e CommandEvent) Quit() {
	e.app.Stop()
}