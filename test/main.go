package main

import (
	"github.com/tdewolff/log"
)

func main() {
	x := 5
	log.Trace("lala", x)
	log.Debug("lala", x)
	log.Info("lala", x)
	log.Warning("lala", x)
	log.Error("lala", x)

	l := log.NewStdLog(log.WarningLevel)
	l.Println("lalala")

	sl := log.NewSlog()
	sl.Warn("lalala")
}
