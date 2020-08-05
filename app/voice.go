package app

import (
	"bufio"
	"encoding/binary"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"io"
	"math"
	"os/exec"
	"strconv"
	"sync"
)

// TODO: Make this thread-safe.

type Voice struct {
	connection *discordgo.VoiceConnection
	pcmIncoming chan *discordgo.Packet
	pcmOutgoing chan []int16

	mutex sync.RWMutex
	outgoingStreams []chan []int16
	mixerCond *sync.Cond
}

func (v *Voice) handleRealtime(event Event) {
	v.pcmIncoming = make(chan *discordgo.Packet, 2)
	v.pcmOutgoing = make(chan []int16, 2)
	v.mixerCond = sync.NewCond(&sync.Mutex{})

	go dgvoice.ReceivePCM(v.connection, v.pcmIncoming)
	go dgvoice.SendPCM(v.connection, v.pcmOutgoing)
	go v.mixer()

	for {
		packet, ok := <- v.pcmIncoming
		if !ok {
			break
		}

		newEvent := Event{
			Kind:        VoiceDataEvent,
			VoicePacket: packet,
			app: event.app,
			guildId: event.guildId,
		}

		event.app.BroadcastEvent(&newEvent)
	}
}

func (v *Voice) mixer() {
	for {
		v.mutex.RLock()
		outgoingStream := v.outgoingStreams
		v.mutex.RUnlock()

		if len(outgoingStream) == 0 {
			v.mixerCond.L.Lock()
			v.mixerCond.Wait()
			v.mixerCond.L.Unlock()
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
			v.pcmOutgoing <- finalFrame
		}
	}
}

func (v *Voice) Play(filename string) chan struct{} {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	stop := make(chan struct{})
	stream := make(chan []int16, 0)

	go func() {
		v.streamFile(filename, stream, stop)

		// Remove stream from outgoing streams.
		v.mutex.Lock()
		for i, outgoingStream := range v.outgoingStreams {
			if outgoingStream == stream {
				v.outgoingStreams = append(v.outgoingStreams[:i], v.outgoingStreams[i+1:]...)
			}
		}
		v.mutex.Unlock()
	}()

	v.outgoingStreams = append(v.outgoingStreams, stream)

	v.mixerCond.L.Lock()
	v.mixerCond.Broadcast()
	v.mixerCond.L.Unlock()

	return stop
}

func (v *Voice) streamFile(filename string, out chan []int16, stop chan struct{}) error {
	frameRate := 48000
	channels := 2
	frameSize := 960

	// Create a shell command "object" to run.
	run := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpegout, err := run.StdoutPipe()
	if err != nil {
		return err
	}

	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)

	// Starts the ffmpeg command
	err = run.Start()
	if err != nil {
		return err
	}

	go func() {
		<-stop
		err = run.Process.Kill()
	}()

	/*
	// Send "speaking" packet over the voice websocket
	err = v.Speaking(true)
	if err != nil {
		OnError("Couldn't set speaking", err)
	}

	// Send not "speaking" packet over the websocket when we finish
	defer func() {
		err := v.Speaking(false)
		if err != nil {
			OnError("Couldn't stop speaking", err)
		}
	}()
	*/

	send := make(chan []int16, 2)
	defer close(send)

	for {
		// read data from ffmpeg stdout
		audiobuf := make([]int16, frameSize*channels)
		err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
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