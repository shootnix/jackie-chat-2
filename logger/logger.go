package logger

import (
	"github.com/shootnix/jackie-chat-2/config"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type Logger struct {
	info  *log.Logger
	debug *log.Logger
	err   *log.Logger
}

var logger *Logger
var once sync.Once

func GetLogger() *Logger {
	once.Do(func() {
		logger = &Logger{}
		cfg := config.GetConfig()
		logger.info = log.New(initHandle(cfg.Logger.Info), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		logger.debug = log.New(initHandle(cfg.Logger.Debug), "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		logger.err = log.New(initHandle(cfg.Logger.Error), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	})
	return logger
}

func (l *Logger) Info(msg string) {
	l.info.Println(msg)
}

func (l *Logger) Debug(msg string) {
	l.debug.Println(msg)
}

func (l *Logger) Error(msg string) {
	l.err.Println(msg)
}

func (l *Logger) Fatal(msg string) {
	l.err.Fatal(msg)
}

func initHandle(out string) io.Writer {
	var h io.Writer

	switch out {
	case "discard":
		h = ioutil.Discard
	case "screen":
		h = os.Stdout
	case "error":
		h = os.Stderr
	default:
		h = openLogFile(out)
	}

	return h
}

func openLogFile(filename string) io.Writer {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err.Error())
	}

	return file
}
