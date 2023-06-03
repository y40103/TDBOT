package utils

import (
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestLog(t *testing.T) {

	logger := Logger{Stdout: true, LogPath: "./test123.log"}

	logger.Init()

	logrus.Infoln("123")
	logrus.Warnln("222222")
	time.Sleep(time.Second * 2)
	logrus.Infoln("123")
	logrus.Warnln("222222")
}
