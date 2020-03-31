package main

import (
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func loadConfig(name string) string {
	token := os.Getenv(name)
	if token != "" {
		return token
	}

	content, err := ioutil.ReadFile(name)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}

func waitForExitSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func formatDate(t time.Time) string {
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

func sendChannelMessage(s *discordgo.Session, channelId string, message string) {
	_, err := s.ChannelMessageSend(channelId, message)
	if err != nil {
		log.Println("Unable to send channel message:", err)
	}
}
