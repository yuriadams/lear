package service

import (
	"log"
	"os"
	"strings"
)

type Logger struct {
	Logger  *log.Logger
	BaseTag string
	Tags    string
}

func NewLogger(baseTag string) *Logger {
	return &Logger{
		Logger:  log.New(os.Stdout, "", log.LstdFlags),
		BaseTag: baseTag,
	}
}

func (l *Logger) SetTags(tags ...string) {
	allTags := append([]string{l.BaseTag}, tags...)
	l.Tags = strings.Join(allTags, " ")
}

func (l *Logger) LogError(baseMsg string, err error) {
	l.Logger.Printf("%s ERROR: %s: %s", l.Tags, baseMsg, err.Error())
}

func (l *Logger) LogInfo(message string) {
	l.Logger.Printf("%s INFO: %s", l.Tags, message)
}
