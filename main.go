package main

import (
	"fmt"
	"net"
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

	/*applDB := &appldbMock{}
	var err error*/

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

type appldbMock struct{}

func (a *appldbMock) AddRoute(pfx net.IPNet, nexthops appldb.Nexthops) error {
	fmt.Printf("Adding Route: %s\n", pfx.String())
	for _, nh := range nexthops {
		fmt.Printf("NH: %s / %s\n", nh.Nexthop.String(), nh.IfName)
	}

	return nil
}

func (a *appldbMock) DelRoute(pfx net.IPNet) error {
	fmt.Printf("Deleting Route: %s\n", pfx.String())
	return nil
}

func (a *appldbMock) Close() error {
	return nil
}
