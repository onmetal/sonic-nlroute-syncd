package routesync

import (
	"net"
	"sync"
	"syscall"

	"github.com/onmetal/sonic-nlroute-syncd/pkg/appldb"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	log "github.com/sirupsen/logrus"
)

type RouteSynchronizer struct {
	appldb *appldb.APPLDB
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func New(appldb *appldb.APPLDB) *RouteSynchronizer {
	return &RouteSynchronizer{
		appldb: appldb,
		stopCh: make(chan struct{}),
	}
}

func (rr *RouteSynchronizer) Start() error {
	routesCh := make(chan netlink.RouteUpdate)
	err := netlink.RouteSubscribe(routesCh, nil)
	if err != nil {
		return errors.Wrap(err, "Unable to subscribe to netlink route updates")
	}

	rr.wg.Add(1)
	go rr.run(routesCh)

	return nil
}

func (rr *RouteSynchronizer) Stop() {
	close(rr.stopCh)
}

func (rr *RouteSynchronizer) StopAndWait() {
	rr.Stop()
	rr.wg.Wait()
}

func (rr *RouteSynchronizer) stopped() bool {
	select {
	case <-rr.stopCh:
		return true
	default:
		return false
	}
}

func (rr *RouteSynchronizer) run(rc chan netlink.RouteUpdate) {
	defer rr.wg.Done()

	for {
		if rr.stopped() {
			return
		}

		u := <-rc

		if u.Dst != nil {
			log.Warning("Ignored route update for non IP destination")
			continue
		}

		switch u.Type {
		case syscall.RTM_NEWROUTE:
			rr.addRoute(&u.Route)
			continue
		case syscall.RTM_DELROUTE:
			rr.delRoute(*u.Dst)
			continue
		}
	}
}

func (rr *RouteSynchronizer) addRoute(r *netlink.Route) {
	nexthops, err := getNexthops(r)
	if err != nil {
		log.WithError(err).Error("Unable to get nexthops")
		return
	}

	err = rr.appldb.AddRoute(*r.Dst, nexthops)
	if err != nil {
		log.WithError(err).Error("Unable to add route")
		return
	}
}

func (rr *RouteSynchronizer) delRoute(dst net.IPNet) {
	err := rr.appldb.DelRoute(dst)
	if err != nil {
		log.WithError(err).Error("Unable to delete route")
		return
	}
}

func getNexthops(r *netlink.Route) (appldb.Nexthops, error) {
	if r.Gw != nil {
		return getNexthopsMonopath(r)
	}

	return getNexthopsMultipath(r)
}

func getNexthopsMonopath(r *netlink.Route) (appldb.Nexthops, error) {
	ifa, err := net.InterfaceByIndex(r.LinkIndex)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get interface by index")
	}

	return appldb.Nexthops{
		{
			Nexthop: r.Gw,
			IfName:  ifa.Name,
		},
	}, nil
}

func getNexthopsMultipath(r *netlink.Route) (appldb.Nexthops, error) {
	nexthops := make(appldb.Nexthops, len(r.MultiPath))

	for i := 0; i < len(r.MultiPath); i++ {
		ifa, err := net.InterfaceByIndex(r.MultiPath[i].LinkIndex)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get interface by index")
		}

		nexthops[i].IfName = ifa.Name
		nexthops[i].Nexthop = r.MultiPath[i].Gw
	}

	return nexthops, nil
}
