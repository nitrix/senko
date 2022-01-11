package helpers

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type Mixer struct {
	inputs []chan []int16
	output chan []int16
	mutex  sync.RWMutex
	cond   *sync.Cond
}

const MixerFrameRate = 48000
const MixerFrameSize = 960
const MixerChannels = 2

func NewMixer() *Mixer {
	m := &Mixer{
		inputs: make([]chan []int16, 0),
		output: make(chan []int16),
		cond:   sync.NewCond(&sync.Mutex{}),
	}

	go m.mixing()

	return m
}

func (m *Mixer) Output() chan []int16 {
	return m.output
}

func (m *Mixer) AddStream(input chan []int16) {
	m.mutex.Lock()
	m.inputs = append(m.inputs, input)
	m.mutex.Unlock()
	m.cond.Broadcast()
}

func (m *Mixer) RemoveStream(stream chan []int16) {
	m.mutex.Lock()
	for i, input := range m.inputs {
		if input == stream {
			m.inputs = append(m.inputs[:i], m.inputs[i+1:]...)
			break
		}
	}
	m.mutex.Unlock()
	m.cond.Broadcast()
}

func (m *Mixer) mixing() {
	for {
		// Wait for inputs to avoid needlessly spinning.
		m.cond.L.Lock()
		for len(m.inputs) == 0 {
			m.cond.Wait()
		}
		m.mutex.RLock()
		inputs := m.inputs
		m.mutex.RUnlock()
		m.cond.L.Unlock()

		outputFrame := make([]int16, MixerFrameSize*MixerChannels)

		for _, input := range inputs {
			select {
			case frame := <-input:
				for i, val := range frame {
					tmp := int32(outputFrame[i]) + int32(val)

					if tmp > math.MaxInt16 {
						tmp = math.MaxInt16
					}

					if tmp < math.MinInt16 {
						tmp = math.MinInt16
					}

					outputFrame[i] = int16(tmp)
				}
			default:
				// Do nothing
			}
		}

		if len(outputFrame) > 0 {
			m.output <- outputFrame
		}
	}
}

func (m *Mixer) playPCM(reader io.Reader) Stopper {
	stopper := NewStopper()
	stream := make(chan []int16)

	m.AddStream(stream)

	go func() {
		for {
			frame := make([]int16, MixerFrameSize*MixerChannels)
			err := binary.Read(reader, binary.LittleEndian, &frame)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			if err != nil {
				break
			}

			stream <- frame
		}

		m.RemoveStream(stream)
		stopper.Stop()
	}()

	return stopper
}

func (m *Mixer) PlayReader(reader io.Reader, normalize bool) (*Stopper, error) {
	args := []string{
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", strconv.Itoa(MixerFrameRate),
		"-ac", strconv.Itoa(MixerChannels),
	}

	if normalize {
		args = append(args, "-filter:a", "loudnorm")
	}

	args = append(args, "pipe:1")

	command := exec.Command("ffmpeg", args...)
	audioOutput, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}

	command.Stdin = reader

	err = command.Start()
	if err != nil {
		return nil, err
	}

	stopper := m.playPCM(audioOutput)

	go func() {
		stopper.Wait()
		command.Process.Kill()
	}()

	return &stopper, nil
}

func (m *Mixer) PlayFile(filename string, normalize bool) (*Stopper, error) {
	// Ensure the file exists.
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, err
	}

	args := []string{
		"-i", filename,
		"-f", "s16le",
		"-ar", strconv.Itoa(MixerFrameRate),
		"-ac", strconv.Itoa(MixerChannels),
	}

	if normalize {
		args = append(args, "-filter:a", "loudnorm")
	}

	args = append(args, "pipe:1")

	command := exec.Command("ffmpeg", args...)
	audioOutput, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = command.Start()
	if err != nil {
		return nil, err
	}

	stopper := m.playPCM(audioOutput)

	go func() {
		stopper.Wait()
		command.Process.Kill()
	}()

	return &stopper, nil
}
