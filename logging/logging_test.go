package logging

import (
	"testing"
	"time"
)

func TestLogging(t *testing.T) {
	Init("./log", "test.log")
	for {
		Debug("debug")
		Debugf("debug: %s", "debug")
		Info("info")
		Warn("warn")
		Error("error")
		time.Sleep(time.Second)
	}
}
