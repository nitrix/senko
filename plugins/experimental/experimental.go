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
	"senko/lib/porcupine"
	"senko/plugins/youtube"
	"strings"
	"time"
)

type Plugin struct {
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
}

func (p *Plugin) Save() error    { return nil }
func (p *Plugin) Restore() error { return nil }

func (p *Plugin) OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) error {
	if strings.HasPrefix(message.Content, "!voice join ") {
		name := strings.TrimPrefix(message.Content, "!voice join ")

		if p.voiceConnection == nil {
			channels, err := session.GuildChannels(message.GuildID)
			if err != nil {
				return err
			}

			for _, channel := range channels {
				if channel.Name == name {
					p.voiceConnection, _ = session.ChannelVoiceJoin(message.GuildID, channel.ID, false, false)
					p.handleRealtimeVoice(session, message.ChannelID)
				}
			}
		}
	}

	if message.Content == "!voice leave" {
		if p.voiceConnection != nil {
			_ = p.voiceConnection.Disconnect()
			p.voiceConnection = nil
		}
	}

	if message.Content == "!sing" {
		if p.voiceConnection != nil {
			mp3Filepath, err := youtube.DownloadAsMp3("https://www.youtube.com/watch?v=K7rEAJn03V0")
			if err != nil {
				return err
			}

			p.interruptPlaying = make(chan bool)
			dgvoice.PlayAudioFile(p.voiceConnection, mp3Filepath, p.interruptPlaying)
		}
	}

	if strings.HasPrefix(message.Content, "!play ") {
		if p.voiceConnection == nil {
			return nil
		}

		url := strings.TrimPrefix(message.Content, "!play ")
		filePath, err := youtube.DownloadAsMp3(url)
		if err != nil {
			return err
		}

		p.interruptPlaying = make(chan bool)
		dgvoice.PlayAudioFile(p.voiceConnection, filePath, p.interruptPlaying)
	}

	if message.Content == "!stop" {
		if p.interruptPlaying != nil {
			close(p.interruptPlaying)
			p.interruptPlaying = nil
		}
	}

	return nil
}

func (p Plugin) handleRealtimeVoice(session *discordgo.Session, channelId string) {
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

	p.buffers = make(map[uint32][]int16)
	p.recordings = make(map[uint32][]int16)
	p.porcupines = make(map[uint32]porcupine.Porcupine)
	p.decoders = make(map[uint32]*gopus.Decoder)

	p.pcmIncoming = make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(p.voiceConnection, p.pcmIncoming)

	p.pcmOutgoing = make(chan []int16, 2)
	go dgvoice.SendPCM(p.voiceConnection, p.pcmOutgoing)

	for {
		packet, ok := <-p.pcmIncoming
		if !ok {
			break
		}

		_, ok = p.buffers[packet.SSRC]
		if !ok {
			p.buffers[packet.SSRC] = make([]int16, 0)
		}

		_, ok = p.recordings[packet.SSRC]
		if !ok {
			p.recordings[packet.SSRC] = make([]int16, 0)
		}

		_, ok = p.porcupines[packet.SSRC]
		if !ok {
			pp, err := porcupine.New("others/model/porcupine_params.pv", kw)
			if err != nil {
				log.Fatalln("Failed to initialize porcupine:", err)
			}
			p.porcupines[packet.SSRC] = pp
		}

		_, ok = p.decoders[packet.SSRC]
		if !ok {
			// Re-decode opus again
			decoder, err := gopus.NewDecoder(16000, 1)
			if err != nil {
				log.Fatalln("Unable to create decoder")
			}
			p.decoders[packet.SSRC] = decoder
		}

		pcm, err := p.decoders[packet.SSRC].Decode(packet.Opus, 960, false)
		if err != nil {
			log.Fatalln("Unable to re decode:", err)
		}

		p.buffers[packet.SSRC] = append(p.buffers[packet.SSRC], pcm...)

		if p.recording {
			p.recordings[packet.SSRC] = append(p.recordings[packet.SSRC], pcm...)
		}

		for len(p.buffers[packet.SSRC]) > porcupine.FrameLength() {
			word, err := p.porcupines[packet.SSRC].Process(p.buffers[packet.SSRC][:porcupine.FrameLength()])
			if err != nil {
				log.Fatalln(err)
			}

			// Detected wake word!
			if word != "" {
				if !p.recording {
					log.Println("Recording START")
					p.recording = true
					p.saidSomething = false
					p.recordings[packet.SSRC] = make([]int16, 0)
				}

				p.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(p.voiceConnection, "others/sounds/attention.mp3", p.interruptPlaying)
			}

			p.buffers[packet.SSRC] = p.buffers[packet.SSRC][porcupine.FrameLength():]
		}

		// p.pcmOutgoing <- packet.PCM // echo

		noise := 0
		for _, v := range pcm {
			absv := math.Abs(float64(v))
			noise += int(absv)
		}

		if p.recording && noise > 10000 {
			if !p.saidSomething {
				log.Println("Said something")
			}

			p.saidSomething = true
			p.lastTimeSaidSomething = time.Now()
		}

		if p.recording && p.saidSomething && time.Since(p.lastTimeSaidSomething) > time.Duration(10 * time.Second) {
			_, _ = session.ChannelMessageSend(channelId, "Giving up")
			p.recording = false
		}

		if p.recording && p.saidSomething && noise < 10000 && time.Since(p.lastTimeSaidSomething) > time.Duration(time.Second) {
			p.recording = false
			log.Println("Recording END")

			p.interruptPlaying = make(chan bool)
			dgvoice.PlayAudioFile(p.voiceConnection, "others/sounds/ok.mp3", p.interruptPlaying)

			err := p.recordingToText(p.recordings[packet.SSRC], session, channelId)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (p Plugin) recordingToText(data []int16, session *discordgo.Session, channelId string) error {
	log.Println("Recording to text...")

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
	// and sample rate information to be transcripted.
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
			_, _ = session.ChannelMessageSend(channelId, fmt.Sprintf("%v (confidence=%3f)\n", alt.Transcript, alt.Confidence))

			if alt.Transcript == "laugh" {
				p.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(p.voiceConnection, "others/sounds/laugh.ogg", p.interruptPlaying)
			}

			if alt.Transcript == "China" {
				p.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(p.voiceConnection, "others/sounds/china.mp3", p.interruptPlaying)
			}

			if alt.Transcript == "Amber" {
				p.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(p.voiceConnection, "others/sounds/ayaya.mp3", p.interruptPlaying)
			}

			if alt.Transcript == "hurt Rica" {
				p.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(p.voiceConnection, "others/sounds/rikka_ow.mp3", p.interruptPlaying)
			}

			if alt.Transcript == "to be continued" {
				p.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(p.voiceConnection, "others/sounds/jojo.mp3", p.interruptPlaying)
			}
		}
	}
	return nil
}