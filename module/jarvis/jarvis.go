package jarvis

import (
	"bytes"
	speech "cloud.google.com/go/speech/apiv1"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"layeh.com/gopus"
	"log"
	"math"
	"runtime"
	"senko/app"
	"senko/lib/porcupine"
	"time"
)

const SoundAttentionFilepath = "others/sounds/attention.mp3"
const SoundOkayFilepath = "others/sounds/click.mp3"

type Jarvis struct {
	// TODO: Cleaning up dangling user data.
	users map[uint32]*UserData
}

type UserData struct {
	buffer []int16
	decoder *gopus.Decoder
	porcupine porcupine.Porcupine
	recording []int16
	isRecording bool
	saidSomething bool
	triggerTime time.Time
	lastSaidSomething time.Time
}

func (j *Jarvis) OnLoad(store *app.Store) {
	j.users = make(map[uint32]*UserData)
}

func (j *Jarvis) OnUnload(store *app.Store) {}

func (j *Jarvis) OnEvent(event *app.Event) error {
	if event.Kind == app.VoiceDataEvent {
		// Create user data if missing.
		userData := j.users[event.VoicePacket.SSRC]
		if userData == nil {
			createdUserData, err := j.createUserData(event.VoicePacket.SSRC)
			if err != nil {
				return err
			}

			userData = createdUserData
		}

		j.processOpus(userData, event)

		err := j.detectWakeWord(userData, event)
		if err != nil {
			return err
		}

		j.detectSilence(userData, event)

		return nil
	}

	return nil

}

func (j *Jarvis) detectSilence(u *UserData, event *app.Event) {
	if u.isRecording && time.Since(u.triggerTime) > 10 * time.Second {
		// Recording for more than 10 seconds is aborted.
		j.resetRecording(u)
	} else if u.saidSomething && time.Since(u.lastSaidSomething) > 1*time.Second {
		// Said something and there's been a silence for the last second.
		event.PlayAudioFile(SoundOkayFilepath)

		text, _ := j.speechToText(u.recording)
		event.DoCommand(text)

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

	pp, err := porcupine.New("others/model/porcupine_params.pv", j.getPorcupineKeyword())
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

func (j *Jarvis) detectWakeWord(u *UserData, event *app.Event) error {
	for len(u.buffer) > porcupine.FrameLength() {
		word, err := u.porcupine.Process(u.buffer[:porcupine.FrameLength()])
		if err != nil {
			return err
		}

		// Detected wake word!
		if word != "" {
			event.PlayAudioFile(SoundAttentionFilepath)

			if !u.isRecording {
				u.isRecording = true
				u.saidSomething = false
				u.recording = make([]int16, 0)
			}

			u.triggerTime = time.Now()
		}

		u.buffer = u.buffer[porcupine.FrameLength():]
	}

	return nil
}

func (j *Jarvis) processOpus(u *UserData, event *app.Event) {
	pcm, err := u.decoder.Decode(event.VoicePacket.Opus, 960, false)
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
		wakeFilepath = "others/wake/senko_windows_2020-06-23_v1.7.0.ppn"
	}

	if runtime.GOOS == "linux" {
		wakeFilepath = "others/wake/senko_linux_2020-06-23_v1.7.0.ppn"
	}

	if runtime.GOOS == "darwin" {
		wakeFilepath = "others/wake/senko_mac_2020-06-23_v1.7.0.ppn"
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