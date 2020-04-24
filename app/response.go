package app

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// This is an abstraction meant to decouple the protocols the bot connects to vs. the internal modules
// issuing commands.

type Response struct {
	session *discordgo.Session
	channelId string
}

type Embed struct {
	Message string

	Title string
	URL string
	Description string
	ThumbnailURL string

	Fields []Field
}

type Field struct {
	Key string
	Value string
}

func (r Response) SendText(content string) error {
	_, err := r.session.ChannelMessageSend(r.channelId, content)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

func (r Response) SendEmbed(embed Embed) error {
	fields := make([]*discordgo.MessageEmbedField, 0)

	for _, field := range embed.Fields {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  field.Key,
			Value: field.Value,
			Inline: true,
		})
	}

	message := discordgo.MessageSend {
		Content: embed.Message,
		Embed: &discordgo.MessageEmbed{
			Title: embed.Title,
			Fields: fields,
			URL:         embed.URL,
			Description: embed.Description,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: embed.ThumbnailURL,
			},
		},
	}

	_, err := r.session.ChannelMessageSendComplex(r.channelId, &message)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

func (r Response) SendFile(path string) error {
	file, err := os.Open(path)
	defer func() {
		_ = file.Close()
	}()

	message := discordgo.MessageSend {
		Files: []*discordgo.File {
			{
				Name:        filepath.Base(path),
				ContentType: mime.TypeByExtension(filepath.Ext(path)),
				Reader:      file,
			},
		},
	}

	_, err = r.session.ChannelMessageSendComplex(r.channelId, &message)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	return nil
}

func (r Response) SendImageFromURL(url string) error {
	buffer := bytes.Buffer{}

	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("unable to download image from URL: %w", err)
	}

	_, err = io.Copy(&buffer, response.Body)
	if err != nil {
		return fmt.Errorf("unable to save image in memory: %w", err)
	}

	message := discordgo.MessageSend {
		Files: []*discordgo.File {
			{
				Name:        filepath.Base(url),
				ContentType: mime.TypeByExtension(response.Header.Get("Content-Type")),
				Reader:      &buffer,
			},
		},
	}

	_, err = r.session.ChannelMessageSendComplex(r.channelId, &message)
	if err != nil {
		return fmt.Errorf("unable to send channel message: %w", err)
	}

	/*
	embed := discordgo.MessageEmbed{
		Image: &discordgo.MessageEmbedImage{
			URL: gif.URL,
		},
	}

	_, err = response.Session.ChannelMessageSendEmbed(response.ChannelId, &embed)
	if err != nil {
		return fmt.Errorf("unable to send message channel: %w", err)
	}
	*/

	return nil
}