package utils

import (
	"bytes"
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

type myFormatter struct {
}

func (self myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// entry 可以理解成 log 的instance , 內含level , msg ....

	var buffer = new(bytes.Buffer) // create 符合 io.writer interface

	timeFormat := entry.Time.Format("2006/01/02 15:04:05.0000")

	//fileval := fmt.Sprintf("From: %v:%v", entry.Caller.File, entry.Caller.Line)
	//fileVal := fmt.Sprintf("%v", self.Title)

	fmt.Fprintf(buffer, "FromFile: %v\n[%v:%v] %v\n[%v]: %v\n\n", entry.Caller.File, entry.Caller.Function, entry.Caller.Line, timeFormat, entry.Level, entry.Message)

	return buffer.Bytes(), nil
}

type Logger struct {
	LogPath string
	Stdout  bool
	Level   logrus.Level
}

func (self *Logger) Init() {

	// set output format
	logrus.SetFormatter(&myFormatter{})
	logrus.SetReportCaller(true)
	// set log level
	if self.Level == 0 {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(self.Level)
	}

	// log rotate
	logger := &lumberjack.Logger{
		Filename:   self.LogPath,
		MaxSize:    100,
		MaxBackups: 10,
		LocalTime:  true,
		MaxAge:     10,
	}
	if self.Stdout == true {
		logrus.SetOutput(io.MultiWriter(logger, os.Stdout))
	} else {
		logrus.SetOutput(io.MultiWriter(logger))
	}

	abslogPath, err := filepath.Abs(self.LogPath)
	if err != nil {
		logrus.Infoln(abslogPath)
	} else {
		logrus.Printf("Set outputLogPath: %v\n", abslogPath)
	}

}
