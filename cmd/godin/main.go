package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/joho/godotenv"
	"wednesday.wtf/godin/api/daemon"
	"wednesday.wtf/godin/api/service"
)

func startServices() {
	for name, service := range daemon.Daemons {
		glog.Infof("Starting daemon: %s", name)

		go service.Start()
	}

	for name, cdnService := range service.CDNServices {
		glog.Infof("Starting CDN service: %s", name)

		go cdnService.Start()
	}
}

func stopServices() {
	for name, service := range daemon.Daemons {
		glog.Infof("Stopping daemon: %s", name)

		service.Stop()
	}

	for name, cdnService := range service.CDNServices {
		glog.Infof("Stopping CDN service: %s", name)

		cdnService.Stop()
	}
}

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
