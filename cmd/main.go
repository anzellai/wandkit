package main

import (
	"flag"
	"os"
	"time"

	"github.com/anzellai/wandkit/libwandkit"
	log "github.com/sirupsen/logrus"
)

var (
	// VERSION build var
	VERSION string
	// COMMIT build var
	COMMIT string
	// BRANCH build var
	BRANCH string
)

var (
	duration time.Duration
	logger   *log.Entry
)

func main() {
	flag.DurationVar(&duration, "duration", time.Duration(time.Second*10), "timeout duration")
	logLevel := flag.Int("logLevel", 0, "set log level")
	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	switch *logLevel {
	case 1:
		log.SetLevel(log.DebugLevel)
	case 2:
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	logger = log.WithFields(log.Fields{
		"app":     "WandKit",
		"version": VERSION,
		"commit":  COMMIT,
		"branch":  BRANCH,
	})

	wk := libwandkit.New(logger, duration)
	wk.Connect()
	wk.Explore()

	select {}
}
