package app

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

type EventCommand struct {
	UserID  UserID
	ChannelID ChannelID
	GuildID GuildID
	Content string
}

func (e EventCommand) Match(pattern string) (map[string]string, bool) {
	matches := make(map[string]string)

	contentParts := strings.Split(e.Content, " ")
	patternParts := strings.Split(pattern, " ")

	for patternIndex, patternPart := range patternParts {
		lastPattern := patternIndex == len(patternParts) - 1

		// When we're still processing patternParts and there are no contentParts left,
		// then it's impossible to get a match.
		if len(contentParts) == 0 {
			return nil, false
		}

		// Placeholder.
		if strings.HasPrefix(patternPart, "<") && strings.HasSuffix(patternPart, ">") {
			name := strings.TrimPrefix(patternPart, "<")
			name = strings.TrimSuffix(name, ">")

			if lastPattern {
				matches[name] = strings.Join(contentParts, " ")
				contentParts = make([]string, 0)
				continue
			} else {
				matches[name] = contentParts[0]
				contentParts = contentParts[1:]
				continue
			}
		}

		// Choice.
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			// TODO
		}

		// Exact match.
		if patternPart == contentParts[0] {
			contentParts = contentParts[1:]
			continue
		}

		return nil, false
	}

	return matches, true
}

func (e EventCommand) Replace(vars map[string]string) {
	contentParts := strings.Split(e.Content, " ")

	for i, part := range contentParts {
		if vars[part] != "" {
			contentParts[i] = vars[part]
		}
	}

	e.Content = strings.Join(contentParts, " ")
}

type EventMessageCreated struct {
	MessageID MessageID
	UserID UserID
	ChannelID ChannelID
	Content string
}

type EventReady struct {
	GuildId GuildID
}

type EventVoiceAlready struct {
	UserID UserID
	ChannelID ChannelID
	GuildID GuildID
}

type EventVoiceData struct {
	UserID UserID
	ChannelID ChannelID
	GuildID GuildID
	VoicePacket *discordgo.Packet // FIXME
}

type EventVoiceJoin struct {
	UserID UserID
	ChannelID ChannelID
	GuildID GuildID
}

type EventVoiceLeave struct {
	UserID UserID
	ChannelID ChannelID
	GuildID GuildID
}
