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
	mutex sync.RWMutex
}

type UserData struct {
	mutex sync.Mutex
	buffer []int16
	decoder *gopus.Decoder
	porcupine *porcupine.Porcupine
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
		j.mutex.RLock()
		userData := j.users[e.VoicePacket.SSRC]
		j.mutex.RUnlock()

		if userData == nil {
			j.mutex.Lock()
			defer j.mutex.Unlock()

			err := j.createUserData(e.VoicePacket.SSRC)
			if err != nil {
				return err
			}

			userData = j.users[e.VoicePacket.SSRC]
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

		log.Println("Speech detected:", text)

		gateway.BroadcastEvent(app.EventCommand{
			// UserID:    event.UserID, FIXME?
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

func (j *Jarvis) createUserData(ssrc uint32) error {
	buffer := make([]int16, 0)

	decoder, err := gopus.NewDecoder(16000, 1)
	if err != nil {
		return errors.New("unable to create opus decoder")
	}

	pp, err := porcupine.New(j.getPorcupineKeyword(), 0.5)
	if err != nil {
		return fmt.Errorf("failed to initialize porcupine: %w", err)
	}

	userData := &UserData{
		buffer: buffer,
		decoder: decoder,
		porcupine: pp,
	}

	j.users[ssrc] = userData

	return nil
}

func (j *Jarvis) detectWakeWord(gateway *app.Gateway, u *UserData, event app.EventVoiceData) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	for len(u.buffer) > u.porcupine.FrameLength() {
		word, err := u.porcupine.Process(u.buffer[:u.porcupine.FrameLength()])
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

		u.buffer = u.buffer[u.porcupine.FrameLength():]
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
		wakeFilepath = "assets/wake/senko_windows_2021-02-04-utc_v1_9_0.ppn"
	}

	if runtime.GOOS == "linux" {
		wakeFilepath = "assets/wake/senko_linux_2021-02-04-utc_v1_9_0.ppn"
	}

	if runtime.GOOS == "darwin" {
		wakeFilepath = "assets/wake/senko_mac_2021-02-04-utc_v1_9_0.ppn"
	}

	return &porcupine.Keyword{
		Label:       "senko",
		FilePath:    wakeFilepath,
		Sensitivity: 0.5,
	}
}

func (j *Jarvis) speechToText(data []int16) (string, error) {
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