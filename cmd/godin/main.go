package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/joho/godotenv"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/server"
)

var serviceStartCallbacks = map[string]func(){
	"server":   server.Start,
	"database": database.Start,
}

var serviceStopCallbacks = map[string]func(){
	"server":   server.Stop,
	"database": database.Stop,
}

// startServices will initialize and start all services that are registered in the serviceStartCallbacks map.
func startServices() {
	for service, start := range serviceStartCallbacks {
		glog.Infof("Starting service: %s", service)

		go start()
	}
}

// stopServices will gracefully shut down any services that need to be stopped before the application exits.
func stopServices() {
	for service, stop := range serviceStopCallbacks {
		glog.Infof("Stopping service: %s", service)

		stop()
	}
}

// awaitShutdownSignal listens for termination signals and blocks until one is received.
func awaitShutdownSignal() {
	killSignal := make(chan os.Signal, 1)

	signal.Notify(killSignal, os.Interrupt, syscall.SIGTERM)

	<-killSignal
}

func init() {
	flag.Parse()
}

func main() {
	if err := godotenv.Load(); err != nil {
		glog.Fatalf("Error loading .env file: %v", err)
	}

	startServices()

	awaitShutdownSignal()

	stopServices()

	glog.Flush()
}
