package app

import (
	"bufio"
	"encoding/binary"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"math"
	"os/exec"
	"strconv"
	"sync"
)

// TODO: Make this thread-safe.

const frameRate = 48000
const channels = 2
const frameSize = 960

type Mixer struct {
	connection *discordgo.VoiceConnection
	pcmIncoming chan *discordgo.Packet
	pcmOutgoing chan []int16

	mutex sync.RWMutex
	outgoingStreams []chan []int16
	mixerCond *sync.Cond
}

func (m *Mixer) handleRealtime(gateway *Gateway, connection *discordgo.VoiceConnection) {
	m.connection = connection
	m.pcmIncoming = make(chan *discordgo.Packet, 2)
	m.pcmOutgoing = make(chan []int16, 2)
	m.mixerCond = sync.NewCond(&sync.Mutex{})

	go dgvoice.ReceivePCM(connection, m.pcmIncoming)
	go dgvoice.SendPCM(connection, m.pcmOutgoing)
	go m.mixer()

	for {
		packet, ok := <- m.pcmIncoming
		if !ok {
			break
		}

		event := EventVoiceData{
			UserID:      UserID(connection.UserID),
			ChannelID:   ChannelID(connection.ChannelID),
			GuildID:     GuildID(connection.GuildID),
			VoicePacket: packet,
		}

		gateway.BroadcastEvent(event)
	}
}

func (m *Mixer) mixer() {
	for {
		m.mutex.RLock()
		outgoingStream := m.outgoingStreams
		m.mutex.RUnlock()

		if len(outgoingStream) == 0 {
			m.mixerCond.L.Lock()
			m.mixerCond.Wait()
			m.mixerCond.L.Unlock()
		}

		finalFrame := make([]int16, 960*2)

		interestingFrame := false

		for _, stream := range outgoingStream {
			select {
			case frame := <- stream:
				for i, val := range frame {
					tmp := int32(finalFrame[i]) + int32(val)

					if tmp > math.MaxInt16 {
						tmp = math.MaxInt16
					}

					if tmp < math.MinInt16 {
						tmp = math.MinInt16
					}

					finalFrame[i] = int16(tmp)
				}
				interestingFrame = true
			default:
				// Do nothing
			}
		}

		if interestingFrame {
			m.pcmOutgoing <- finalFrame
		}
	}
}

func (m *Mixer) PlayFile(filename string) chan struct{} {
	// Create a shell command "object" to run.
	run := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpegout, err := run.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// Starts the ffmpeg command
	err = run.Start()
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	stop := m.PlayReader(ffmpegout)

	go func() {
		<-stop
		err = run.Process.Kill()
	}()

	return stop
}

func (m *Mixer) PlayReader(reader io.Reader) chan struct{} {
	stop := make(chan struct{})
	stream := make(chan []int16, 0)

	buffer := bufio.NewReaderSize(reader, 16384)

	go func() {
		err := m.streamIt(buffer, stream, stop)
		if err != nil {
			log.Println(err)
		}

		m.removeOutgoingStream(stream)
	}()

	m.addOutgoingStream(stream)

	return stop
}

func (m *Mixer) wakeUpMixer() {
	m.mixerCond.L.Lock()
	m.mixerCond.Broadcast()
	m.mixerCond.L.Unlock()
}

func (m *Mixer) addOutgoingStream(stream chan []int16) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.outgoingStreams = append(m.outgoingStreams, stream)

	m.wakeUpMixer()
}

func (m *Mixer) removeOutgoingStream(stream chan []int16) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, outgoingStream := range m.outgoingStreams {
		if outgoingStream == stream {
			m.outgoingStreams = append(m.outgoingStreams[:i], m.outgoingStreams[i+1:]...)
		}
	}

	m.wakeUpMixer()
}

func (m *Mixer) streamIt(reader io.Reader, out chan []int16, stop chan struct{}) error {
	// TODO: Stop streaming when stop gets closed

	for {
		// read data from ffmpeg stdout
		audiobuf := make([]int16, frameSize*channels)
		err := binary.Read(reader, binary.LittleEndian, &audiobuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Send received PCM to the sendPCM channel
		out <- audiobuf
	}
}
