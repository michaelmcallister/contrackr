package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/michaelmcallister/contrackr/pkg/contrackr/engine"

	log "github.com/golang/glog"
)

var captureInterface string

func init() {
	const (
		defaultIface = "eth0"
		usage        = "network interface to capture"
	)
	flag.StringVar(&captureInterface, "interface", defaultIface, usage)
	flag.StringVar(&captureInterface, "i", defaultIface, usage)
}

func main() {
	flag.Parse()
	eng, err := engine.New(captureInterface)
	if err != nil {
		log.Exit(err)
	}

	// Capture SIGINT, SIGTERM and run some cleanup.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("Shutting down...")
		if err := eng.Close(); err != nil {
			log.Warning("Error when closing: ", err)
		}
		log.Info("Goodbye!")
		os.Exit(1)
	}()

	log.Info("Running...")
	eng.Run()
}
