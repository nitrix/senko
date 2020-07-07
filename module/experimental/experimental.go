package experimental

import (
	bytes "bytes"
	speech "cloud.google.com/go/speech/apiv1"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"layeh.com/gopus"
	"log"
	"math"
	"runtime"
	"senko/app"
	"senko/lib/porcupine"
	"senko/module/youtube"
	"strings"
	"time"
)

type Experimental struct {
	voiceConnection  *discordgo.VoiceConnection
	pcmIncoming      chan *discordgo.Packet
	pcmOutgoing      chan []int16
	interruptPlaying chan bool
	recording bool
	saidSomething bool
	lastTimeSaidSomething time.Time

	recordings map[uint32][]int16

	buffers    map[uint32][]int16
	decoders   map[uint32]*gopus.Decoder
	porcupines map[uint32]porcupine.Porcupine

	app *app.App
}

func (e *Experimental) Load() error { return nil }

func (e *Experimental) Unload() error { return nil }

func (e *Experimental) OnCommand(event *app.CommandEvent) error {
	if strings.HasPrefix(event.Content, "voice join ") {
		name := strings.TrimPrefix(event.Content, "voice join ")

		var err error

		if e.voiceConnection == nil {
			e.voiceConnection, err = event.JoinVoice(name)
			if err != nil {
				return err
			}

			e.handleRealtimeVoice(event)
		}
	}

	if event.Content == "voice leave" {
		if e.voiceConnection != nil {
			_ = e.voiceConnection.Disconnect()
			e.voiceConnection = nil
		}
	}

	if event.Content == "sing" {
		if e.voiceConnection != nil {
			mp3Filepath, err := youtube.DownloadAsMp3("https://www.youtube.com/watch?v=K7rEAJn03V0")
			if err != nil {
				return err
			}

			e.interruptPlaying = make(chan bool)
			dgvoice.PlayAudioFile(e.voiceConnection, mp3Filepath, e.interruptPlaying)
		}
	}

	if strings.HasPrefix(event.Content, "play ") {
		if e.voiceConnection == nil {
			return nil
		}

		what := strings.TrimPrefix(event.Content, "play ")

		if !strings.HasPrefix(what, "http") {
			what = fmt.Sprintf("ytsearch1:%s", what)
		}

		filePath, err := youtube.DownloadAsMp3(what)
		if err != nil {
			return err
		}

		e.interruptPlaying = make(chan bool)
		dgvoice.PlayAudioFile(e.voiceConnection, filePath, e.interruptPlaying)
	}

	if event.Content == "stop" {
		if e.interruptPlaying != nil {
			close(e.interruptPlaying)
			e.interruptPlaying = nil
		}
	}

	return nil
}

func (e Experimental) OnMessageCreated(event *app.MessageCreatedEvent) error { return nil }

func (e Experimental) handleRealtimeVoice(event *app.CommandEvent) {
	wake := ""
	if runtime.GOOS == "windows" {
		wake = "others/wake/senko_windows_2020-06-23_v1.7.0.ppn"
	}

	if runtime.GOOS == "linux" {
		wake = "others/wake/senko_linux_2020-06-23_v1.7.0.ppn"
	}

	if runtime.GOOS == "darwin" {
		wake = "others/wake/senko_mac_2020-06-23_v1.7.0.ppn"
	}

	kw := &porcupine.Keyword{
		Label:       "senko",
		FilePath:    wake,
		Sensitivity: 0.5,
	}

	e.buffers = make(map[uint32][]int16)
	e.recordings = make(map[uint32][]int16)
	e.porcupines = make(map[uint32]porcupine.Porcupine)
	e.decoders = make(map[uint32]*gopus.Decoder)

	e.pcmIncoming = make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(e.voiceConnection, e.pcmIncoming)

	e.pcmOutgoing = make(chan []int16, 2)
	go dgvoice.SendPCM(e.voiceConnection, e.pcmOutgoing)

	for {
		packet, ok := <-e.pcmIncoming
		if !ok {
			break
		}

		_, ok = e.buffers[packet.SSRC]
		if !ok {
			e.buffers[packet.SSRC] = make([]int16, 0)
		}

		_, ok = e.recordings[packet.SSRC]
		if !ok {
			e.recordings[packet.SSRC] = make([]int16, 0)
		}

		_, ok = e.porcupines[packet.SSRC]
		if !ok {
			pp, err := porcupine.New("others/model/porcupine_params.pv", kw)
			if err != nil {
				log.Fatalln("Failed to initialize porcupine:", err)
			}
			e.porcupines[packet.SSRC] = pp
		}

		_, ok = e.decoders[packet.SSRC]
		if !ok {
			// Re-decode opus again
			decoder, err := gopus.NewDecoder(16000, 1)
			if err != nil {
				log.Fatalln("Unable to create decoder")
			}
			e.decoders[packet.SSRC] = decoder
		}

		pcm, err := e.decoders[packet.SSRC].Decode(packet.Opus, 960, false)
		if err != nil {
			log.Fatalln("Unable to re decode:", err)
		}

		e.buffers[packet.SSRC] = append(e.buffers[packet.SSRC], pcm...)

		if e.recording {
			e.recordings[packet.SSRC] = append(e.recordings[packet.SSRC], pcm...)
		}

		for len(e.buffers[packet.SSRC]) > porcupine.FrameLength() {
			word, err := e.porcupines[packet.SSRC].Process(e.buffers[packet.SSRC][:porcupine.FrameLength()])
			if err != nil {
				log.Fatalln(err)
			}

			// Detected wake word!
			if word != "" {
				if !e.recording {
					log.Println("Recording START")
					e.recording = true
					e.saidSomething = false
					e.recordings[packet.SSRC] = make([]int16, 0)
				}

				// FIXME: Interrupt audio until I have a proper mixer.
				if e.interruptPlaying != nil {
					close(e.interruptPlaying)
					e.interruptPlaying = nil
				}

				e.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(e.voiceConnection, "others/sounds/attention.mp3", e.interruptPlaying)
			}

			e.buffers[packet.SSRC] = e.buffers[packet.SSRC][porcupine.FrameLength():]
		}

		// e.pcmOutgoing <- packet.PCM // echo

		noise := 0
		for _, v := range pcm {
			absv := math.Abs(float64(v))
			noise += int(absv)
		}

		if e.recording && noise > 10000 {
			if !e.saidSomething {
				log.Println("Said something")
			}

			e.saidSomething = true
			e.lastTimeSaidSomething = time.Now()
		}

		if e.recording && e.saidSomething && time.Since(e.lastTimeSaidSomething) > 10 * time.Second {
			e.recording = false
		}

		if e.recording && e.saidSomething && noise < 10000 && time.Since(e.lastTimeSaidSomething) > time.Second {
			e.recording = false
			log.Println("Recording END")

			err := e.speechToText(event, e.recordings[packet.SSRC])
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (e Experimental) speechToText(event *app.CommandEvent, data []int16) error {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	if err != nil {
		return err
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
		return err
	}

	// Print the results.
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			// FIXME error handling
			formatted := fmt.Sprintf("%v (confidence=%3f)\n", alt.Transcript, alt.Confidence)

			log.Println(formatted)
			_ = event.Reply(formatted)

			go func(command string) {
				_ = event.DoCommand(command)
			}(alt.Transcript)
		}
	}

	return nil
}