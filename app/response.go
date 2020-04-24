package app

import "github.com/bwmarrin/discordgo"

// FIXME
// Only works for Discord right now, but is meant to become a request/response model for any protocl
// the bot connects to. (XMPP, IRC, SMS, etc).

type Response struct {
	Session *discordgo.Session // FIXME: private
	ChannelId string // FIXME: private
}
