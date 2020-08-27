package jarvis

import (
	"bytes"
	speech "cloud.google.com/go/speech/apiv1"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nitrix/porcupine"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"layeh.com/gopus"
	"log"
	"math"
	"runtime"
	"senko/app"
	"sync"
	"time"
)

const SoundAttentionFilepath = "assets/sounds/attention.mp3"
const SoundOkayFilepath = "assets/sounds/click.mp3"

type Jarvis struct {
	// TODO: Cleaning up dangling user data.
	users map[uint32]*UserData
}

type UserData struct {
	mutex sync.Mutex
	buffer []int16
	decoder *gopus.Decoder
	porcupine porcupine.Porcupine
	recording []int16
	isRecording bool
	saidSomething bool
	triggerTime time.Time
	lastSaidSomething time.Time
}

func (j *Jarvis) OnRegister(store *app.Store) {
	j.users = make(map[uint32]*UserData)
}

func (j *Jarvis) OnEvent(gateway *app.Gateway, event interface{}) error {
	switch e := event.(type) {
	case app.EventVoiceData:
		// Create user data if missing.
		userData := j.users[e.VoicePacket.SSRC]
		if userData == nil {
			createdUserData, err := j.createUserData(e.VoicePacket.SSRC)
			if err != nil {
				return err
			}

			userData = createdUserData
		}

		j.processOpus(userData, e.VoicePacket)

		err := j.detectWakeWord(gateway, userData, e)
		if err != nil {
			return err
		}

		j.detectSilence(gateway, userData, e)

		return nil
	}

	return nil
}

func (j *Jarvis) detectSilence(gateway *app.Gateway, u *UserData, event app.EventVoiceData) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.isRecording && time.Since(u.triggerTime) > 10 * time.Second {
		// Recording for more than 10 seconds is aborted.
		j.resetRecording(u)
	} else if u.saidSomething && time.Since(u.lastSaidSomething) > time.Second {
		// Said something and there's been a silence for the last second.
		_, _ = gateway.PlayAudioFile(event.GuildID, SoundOkayFilepath)

		text, err := j.speechToText(u.recording)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("Voice command:", text)

		gateway.BroadcastEvent(app.EventCommand{
			UserID:    event.UserID,
			ChannelID: event.ChannelID,
			GuildID:   event.GuildID,
			Content:   text,
		})

		j.resetRecording(u)
	}
}

func (j *Jarvis) resetRecording(u *UserData) {
	u.saidSomething = false
	u.isRecording = false
	u.recording = make([]int16, 0)
}

func (j *Jarvis) createUserData(ssrc uint32) (*UserData, error) {
	buffer := make([]int16, 0)

	decoder, err := gopus.NewDecoder(16000, 1)
	if err != nil {
		return nil, errors.New("unable to create opus decoder")
	}

	pp, err := porcupine.New("assets/model/porcupine_params.pv", j.getPorcupineKeyword())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize porcupine: %w", err)
	}

	userData := &UserData{
		buffer: buffer,
		decoder: decoder,
		porcupine: pp,
	}

	j.users[ssrc] = userData

	return userData, nil
}

func (j *Jarvis) detectWakeWord(gateway *app.Gateway, u *UserData, event app.EventVoiceData) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	for len(u.buffer) > porcupine.FrameLength() {
		word, err := u.porcupine.Process(u.buffer[:porcupine.FrameLength()])
		if err != nil {
			return err
		}

		// Detected wake word!
		if word != "" {
			_, _ = gateway.PlayAudioFile(event.GuildID, SoundAttentionFilepath)

			if !u.isRecording {
				u.isRecording = true
				u.saidSomething = false
				u.recording = make([]int16, 0)

				go func() {
					for i := 0; i < 10; i++ {
						j.detectSilence(gateway, u, event)
						time.Sleep(500 * time.Millisecond)
					}
				}()
			}

			u.triggerTime = time.Now()
		}

		u.buffer = u.buffer[porcupine.FrameLength():]
	}

	return nil
}

func (j *Jarvis) processOpus(u *UserData, voicePacket *discordgo.Packet) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	pcm, err := u.decoder.Decode(voicePacket.Opus, 960, false)
	if err != nil {
		log.Println("Unable to re decode:", err)
	}

	u.buffer = append(u.buffer, pcm...)

	if u.isRecording {
		u.recording = append(u.recording, pcm...)
	}

	if u.isRecording {
		noise := 0
		for _, val := range pcm {
			noise += int(math.Abs(float64(val)))
		}
		if noise > 100000 {
			u.saidSomething = true
			u.lastSaidSomething = time.Now()
		}
	}
}

func (j *Jarvis) getPorcupineKeyword() *porcupine.Keyword {
	wakeFilepath := ""

	if runtime.GOOS == "windows" {
		wakeFilepath = "assets/wake/senko_windows_2020-06-23_v1.7.0.ppn"
	}

	if runtime.GOOS == "linux" {
		wakeFilepath = "assets/wake/senko_linux_2020-06-23_v1.7.0.ppn"
	}

	if runtime.GOOS == "darwin" {
		wakeFilepath = "assets/wake/senko_mac_2020-06-23_v1.7.0.ppn"
	}

	return &porcupine.Keyword{
		Label:       "senko",
		FilePath:    wakeFilepath,
		Sensitivity: 0.5,
	}
}

func (j Jarvis) speechToText(data []int16) (string, error) {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	if err != nil {
		return "", err
	}

	theBytes := bytes.Buffer{}
	for _, d := range data {
		_ = binary.Write(&theBytes, binary.LittleEndian, d)
	}

	// Send the contents of the audio file with the encoding and
	// and sample rate information to be transcribed.
	resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: 16000,
			LanguageCode:    "en-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: theBytes.Bytes()},
		},
	})

	// TODO: check resp error

	if err != nil {
		return "", err
	}

	// Print the results.
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			// FIXME error handling
			// alt.Confidence
			return alt.Transcript, nil
		}
	}

	return "", nil
}