package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelmcallister/contrackr/pkg/contrackr/engine"

	log "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	captureInterface string
	metricsAddr      string
)

var (
	connectionsTracked = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "contrackr_tracked_connections",
		Help: "The current number of tracked requests",
	})
)

func init() {
	const (
		defaultIface = "eth0"
		ifaceUsage   = "network interface to capture"

		defaultMetricsAddr = ":2112"
		metricsUsage       = "the addr to listen on for metrics"
	)
	flag.StringVar(&captureInterface, "interface", defaultIface, ifaceUsage)
	flag.StringVar(&captureInterface, "i", defaultIface, ifaceUsage)
	flag.StringVar(&metricsAddr, "port", defaultMetricsAddr, metricsUsage)
	flag.StringVar(&metricsAddr, "p", defaultMetricsAddr, metricsUsage)
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

	go func() {
		for {
			st := eng.Stats()
			connectionsTracked.Set(float64(st.TotalConnections))
			time.Sleep(2 * time.Second)
		}
	}()

	log.Info("Running...")
	go eng.Run()
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(metricsAddr, nil); err != nil {
		log.Error("unable to serve metrics handler: %v", err)
	}
}
