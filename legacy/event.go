package legacy

/*

import (
	"github.com/bwmarrin/discordgo"
	"senko/app"
)

type Event struct {
	Kind string
	User app.User

	// TODO: Ideally this would be more abstracted and not depend on Discord (also private).
	Content string

	channelId string
	guildId string
	app *app.App
	session     *discordgo.Session
	message     *discordgo.MessageCreate
	VoicePacket *discordgo.Packet
}

func (e Event) IsChannelInUse(channelId string) (bool, error) {
	e.app.mutex.Lock()
	defer e.app.mutex.Unlock()

	for _, currentChannelId := range e.app.currentVoiceChannel {
		if currentChannelId == channelId {
			return true, nil
		}
	}

	return false, nil
}

func (e Event) React(emoji string) error {
	return e.session.MessageReactionAdd(e.message.ChannelID, e.message.ID, emoji)
}

func (e Event) ReplyComplex(msg discordgo.MessageSend) error {
	if e.channelId == "" {
		return errors.New("replying to a voice command unsupported") // FIXME
	}

	_, err := e.session.ChannelMessageSendComplex(e.channelId, &msg)
	if err != nil {
		return err
	}

	return nil
}

func (e Event) ResolveNick(nick string) (User, error) {
	members, err := e.session.GuildMembers(e.guildId, "", 1000)
	if err != nil {
		return User{}, err
	}

	for _, member := range members {
		if member.User.Username == nick || member.Nick == nick {
			return User{
				gatewayName: discord.Discord{}.Name(),
				identifier: member.User.ID,
			}, nil
		}
	}

	return User{}, errors.New("nick not found")
}

func (e Event) ReplyFile(path string) error {
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

func (e Event) PlayAudioFile(filepath string) (chan struct{}, error) {
	voice := e.app.voices[e.guildId]
	if voice == nil {
		return nil, errors.New("no active voice connection available")
	}

	stop := voice.Play(filepath)

	return stop, nil
}

func (e Event) ReplyEmbed(embed discordgo.MessageEmbed) error {
	_, err := e.session.ChannelMessageSendEmbed(e.channelId, &embed)
	if err != nil {
		return err
	}

	return nil
}

func (e Event) FindChannelByName(name string) (*discordgo.Channel, error) {
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

func (e Event) DoCommand(command string) {
	e.Kind = CommandEvent
	e.Content = command

	for _, module := range e.app.modules {
		err := module.OnRequest(&e)
		if err != nil {
			log.Println(err)
		}
	}
}

func (e Event) JoinVoice(channelId string) error {
	var connection *discordgo.VoiceConnection

	e.app.mutex.Lock()
	defer e.app.mutex.Unlock()

	voice := e.app.voices[e.guildId]
	if voice != nil {
		connection = voice.connection
	}

	// Leave the previous channel if we were in one and it isn't the same as what is requested.
	if connection != nil && connection.ChannelID != channelId {
		e.app.mutex.Unlock()
		err := e.LeaveVoice(connection.ChannelID)
		e.app.mutex.Lock()
		if err != nil {
			return err
		}
	}

	if voice == nil {
		voice = &Voice{}
		e.app.voices[e.guildId] = voice
	}

	connection, err := e.session.ChannelVoiceJoin(e.guildId, channelId, false, false)
	if err != nil {
		return err
	}

	voice.connection = connection

	go voice.handleRealtime(e)

	return nil
}

func (e Event) LeaveVoice(channelId string) error {
	e.app.mutex.Lock()
	defer e.app.mutex.Unlock()

	voice := e.app.voices[e.guildId]

	if voice == nil || voice.connection == nil {
		return nil
	}

	if voice.connection.ChannelID != channelId {
		return nil
	}

	delete(e.app.voices, e.guildId)

	err := voice.connection.Disconnect()
	if err != nil {
		return err
	}

	return nil
}

func (e Event) CurrentVoiceChannelId() string {
	return e.app.voices[e.guildId].connection.ChannelID
}

func (e Event) Quit() error {
	e.app.Stop()
	return nil
}
*/
