package infrastructure

import (
	"fmt"
	"log"
)

type LoggerStack interface {
	WriteAll(message string)
}

type loggerStack struct {
	channels []string
}

func NewLoggerStack(channels []string) LoggerStack {
	return &loggerStack{channels: channels}
}

func (l *loggerStack) WriteAll(message string) {
	for _, channel := range l.channels {
		switch channel {
		case "stdout":
			log.Println(message)
		default:
			fmt.Println(message)
		}
	}
}
