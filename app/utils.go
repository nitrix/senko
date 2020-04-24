package app

import (
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func GetToken(name string) string {
	token := os.Getenv(name)
	if token != "" {
		return token
	}

	content, err := ioutil.ReadFile(rootPath() + "/config/" + name)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}

func rootPath() string {
	_, fileName, _, _ := runtime.Caller(0)
	return filepath.ToSlash(filepath.Dir(fileName)) + "/../"
}

func waitForExitSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
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
