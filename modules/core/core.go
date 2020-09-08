package core

import (
	"cloud.google.com/go/texttospeech/apiv1"
	"context"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
	"io/ioutil"
	"os"
	"senko/app"
)

type Core struct{
	stops map[string]chan struct{}
}

func (c *Core) OnRegister(store *app.Store) {
	c.stops = make(map[string]chan struct{})
}

func (c *Core) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventCommand:
		if _, ok := e.Match("help"); ok {
			return gateway.SendMessage(e.ChannelID, "For a list of commands and their usage, visit https://github.com/nitrix/senko/blob/master/docs/commands.md")
		}

		if _, ok := e.Match("quit"); ok {
			gateway.Stop()
			return nil
		}

		if vars, ok := e.Match("voice join <channel>"); ok {
			channelId, err := gateway.FindChannelByName(e.GuildID, vars["channel"])
			if err != nil {
				return err
			}

			return gateway.JoinVoice(e.GuildID, channelId)
		}

		if _, ok := e.Match("voice leave"); ok {
			return gateway.LeaveVoiceAny(e.GuildID)
		}

		if vars, ok := e.Match("say <text>"); ok {
			return c.say(gateway, e.GuildID, vars["text"])
		}
	}

	return nil
}

func (c *Core) say(gateway *app.Gateway, guildId app.GuildID, what string) error {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return err
	}

	request := texttospeechpb.SynthesizeSpeechRequest{
		AudioConfig: &texttospeechpb.AudioConfig{
			// I wish we would use LINEAR16 here and PlayAudioStream but something with the sampling rate is wrong.
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: what,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode:    "en-US",
			Name:            "en-US-Wavenet-H",
			SsmlGender:      texttospeechpb.SsmlVoiceGender_FEMALE,
		},
	}

	response, err := client.SynthesizeSpeech(ctx, &request)
	if err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile("", "voice")
	if err != nil {
		return err
	}

	_, err = tmpFile.Write(response.AudioContent)
	if err != nil {
		return err
	}

	err = tmpFile.Close()
	if err != nil {
		return err
	}

	stopper, err := gateway.PlayAudioFile(guildId, tmpFile.Name())
	if err != nil {
		return err
	}

	// Can't remove the temporary file until it's been played.
	go func() {
		<- stopper
		_ = os.Remove(tmpFile.Name())
	}()

	return err
}
