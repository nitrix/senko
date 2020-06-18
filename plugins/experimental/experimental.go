package experimental

import (
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
	"log"
	"runtime"
	"senko/lib/porcupine"
	"senko/plugins/youtube"
	"strings"
)

type Plugin struct {
	voiceConnection *discordgo.VoiceConnection
	pcmIncoming chan *discordgo.Packet
	pcmOutgoing chan []int16
	interruptPlaying chan bool

	buffers map[uint32][]int16
	decoders map[uint32]*gopus.Decoder
	porcupines map[uint32]porcupine.Porcupine
}

func (p *Plugin) Save() error { return nil }
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
					p.handleRealtimeVoice()
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

func (p Plugin) handleRealtimeVoice() {
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
		Value: "senko",
		FilePath: wake,
		Sensitivity: 0.5,
	}

	p.buffers = make(map[uint32][]int16)
	p.porcupines = make(map[uint32]porcupine.Porcupine)
	p.decoders = make(map[uint32]*gopus.Decoder)

	p.pcmIncoming = make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(p.voiceConnection, p.pcmIncoming)

	p.pcmOutgoing = make(chan []int16, 2)
	go dgvoice.SendPCM(p.voiceConnection, p.pcmOutgoing)

	_ = p.voiceConnection.Speaking(true)
	for {
		packet, ok := <-p.pcmIncoming
		if !ok {
			break
		}

		_, ok = p.buffers[packet.SSRC]
		if !ok {
			p.buffers[packet.SSRC] = make([]int16, 0)
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

		for ; len(p.buffers[packet.SSRC]) > porcupine.FrameLength() ; {
			word, err := p.porcupines[packet.SSRC].Process(p.buffers[packet.SSRC][:porcupine.FrameLength()])
			if err != nil {
				log.Fatalln(err)
			}

			if word != "" {
				log.Println("Detected wake word!")
				p.interruptPlaying = make(chan bool)
				dgvoice.PlayAudioFile(p.voiceConnection, "others/att/att1.mp3", p.interruptPlaying)
			}

			p.buffers[packet.SSRC] = p.buffers[packet.SSRC][porcupine.FrameLength():]
		}

		// p.pcmOutgoing <- packet.PCM // echo
	}
	_ = p.voiceConnection.Speaking(false)
}
