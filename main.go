package main

import (
	"github.com/onmetal/sonic-nlroute-syncd/pkg/appldb"
	"github.com/onmetal/sonic-nlroute-syncd/pkg/routesync"

	log "github.com/sirupsen/logrus"
)

func main() {
	applDB := appldb.New()
	rtSync := routesync.New(applDB)

	err := rtSync.Start()
	if err != nil {
		log.WithError(err).Panic("Unable to start route synchronizer")
	}

	select {}
}
