package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/onmetal/sonic-nlroute-syncd/pkg/appldb"
	"github.com/onmetal/sonic-nlroute-syncd/pkg/routesync"

	log "github.com/sirupsen/logrus"
)

func main() {
	applDB := appldb.New()
	err := applDB.Test()
	if err != nil {
		log.WithError(err).Fatal("Connection to APPL_DB failed")
	}

	rtSync := routesync.New(applDB)
	err = rtSync.Start()
	if err != nil {
		log.WithError(err).Fatal("Unable to start route synchronizer")
	}

	sigs := make(chan os.Signal, 0)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	<-sigs
	rtSync.StopAndWait()

	err = applDB.Close()
	if err != nil {
		log.WithError(err).Fatal("Unable to close Redis connection")
	}
}
