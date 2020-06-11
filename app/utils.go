package app

import (
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func GetToken(name string) string {
	token := os.Getenv(name)
	if token != "" {
		return token
	}

	_, fileName, _, _ := runtime.Caller(0)
	rootPath := filepath.ToSlash(filepath.Dir(fileName)) + "/../"

	content, err := ioutil.ReadFile(rootPath + "/config/" + name)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}

	suffix := "th"

	switch t.Day() {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}

	return t.Format("January 2" + suffix + " 2006")
}

func DiscordSendFile(session *discordgo.Session, channelId string, path string) error {
	file, err := os.Open(path)
	defer func() {
		_ = file.Close()
	}()

	message := discordgo.MessageSend {
		Files: []*discordgo.File {
			{
				Name:        filepath.Base(path),
				ContentType: mime.TypeByExtension(filepath.Ext(path)),
				Reader:      file,
			},
		},
	}

	_, err = session.ChannelMessageSendComplex(channelId, &message)

	return err
}