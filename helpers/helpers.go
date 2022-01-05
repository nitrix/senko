package helpers

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

func FindChannelByName(s *discordgo.Session, guildID string, channelName string) (*discordgo.Channel, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get channels: %w", err)
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel, nil
		}
	}

	return nil, errors.New("unable to find channel")
}

func MatchWithNamedParams(rule string, haystack string) (map[string]string, bool) {
	// Disasemble the rule into parts, turn the parts into valid regex capture groups,
	// then we can reassemble it all into a proper complete regex.
	ruleParts := strings.Split(rule, " ")
	for i, rulePart := range ruleParts {
		last := i == len(ruleParts)-1

		if strings.HasPrefix(rulePart, "<") && strings.HasSuffix(rulePart, ">") {
			rulePart = strings.TrimPrefix(rulePart, "<")
			rulePart = strings.TrimSuffix(rulePart, ">")

			if last {
				rulePart = "(?P<" + rulePart + ">(.*))"
			} else {
				rulePart = "(?P<" + rulePart + ">(\\S*))"
			}

			ruleParts[i] = rulePart
		}
	}
	strRegex := "^" + strings.Join(ruleParts, " ") + "$"

	// Process the regex.
	regex := regexp.MustCompile(strRegex)
	match := regex.FindStringSubmatch(haystack)

	// Naming the capture groups in the output.
	output := make(map[string]string)
	nameCount := 0

	for i, name := range regex.SubexpNames() {
		if name != "" {
			nameCount++

			if i > 0 && i <= len(match) {
				output[name] = match[i]
			}
		}
	}

	ok := match != nil && nameCount == len(output)

	// The output.
	return output, ok
}

func PlayFile(voiceConnection *discordgo.VoiceConnection, filename string, normalize bool) (chan struct{}, error) {
	const frameRate = 48000
	const frameSize = 960
	const channels = 2

	// pcmIncoming := make(chan *discordgo.Packet)
	// go dgvoice.ReceivePCM(voiceConnection, pcmIncoming)

	pcmOutgoing := make(chan []int16)

	go func() {
		dgvoice.SendPCM(voiceConnection, pcmOutgoing)
	}()

	// Ensure the file exists.
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, err
	}

	args := []string{
		"-i", filename,
		"-f", "s16le",
		"-ar", strconv.Itoa(frameRate),
		"-ac", strconv.Itoa(channels),
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

	stopper := make(chan struct{})
	audioBuffer := bufio.NewReaderSize(audioOutput, 16384)

	go func() {
		<-stopper
		err = command.Process.Kill()
	}()

	go func() {
		for {
			tmp := make([]int16, frameSize*channels)
			err := binary.Read(audioBuffer, binary.LittleEndian, &tmp)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				close(stopper)
				break
			}
			if err != nil {
				close(stopper)
				break
			}

			pcmOutgoing <- tmp
		}

		close(pcmOutgoing) // FIXME: Re-visit this, it writes to stderr.
	}()

	return stopper, nil
}
