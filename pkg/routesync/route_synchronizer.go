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

// RouteSynchronizer consumes Netlink route messages and synchronizes them into the APPL_DB
type RouteSynchronizer struct {
	appldb         *appldb.APPLDB
	rc             chan netlink.RouteUpdate
	stopCh         chan struct{}
	wg             sync.WaitGroup
	ifNameResolver ifNameResolver
}

// New creates a new RouteSynchronizer
func New(appldb *appldb.APPLDB) *RouteSynchronizer {
	return &RouteSynchronizer{
		rc:             make(chan netlink.RouteUpdate),
		appldb:         appldb,
		stopCh:         make(chan struct{}),
		ifNameResolver: &ifNameResolverNetlink{},
	}
}

// Start starts the synchronizer
func (rr *RouteSynchronizer) Start() error {
	err := netlink.RouteSubscribe(rr.rc, rr.stopCh)
	if err != nil {
		return errors.Wrap(err, "Unable to subscribe to netlink route updates")
	}

	rr.wg.Add(1)
	go rr.run()

	return nil
}

// Stop stops the synchronizer and doesn't wait for it to actually stop
func (rr *RouteSynchronizer) Stop() {
	close(rr.stopCh)
	close(rr.rc)
}

// StopAndWait stops the synchronizer and waits for it to actually stop
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

func (rr *RouteSynchronizer) run() {
	defer rr.wg.Done()

	for {
		if rr.stopped() {
			return
		}

		u := <-rr.rc

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
	nexthops, err := rr.getNexthops(r)
	if err != nil {
		log.WithError(err).Error("Unable to get nexthops")
		return
	}

	if nexthops == nil {
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

func (rr *RouteSynchronizer) getNexthops(r *netlink.Route) (appldb.Nexthops, error) {
	if r.Gw == nil && r.MultiPath == nil {
		return nil, nil
	}

	if r.Gw != nil {
		return rr.getNexthopsMonopath(r)
	}

	return rr.getNexthopsMultipath(r)
}

func (rr *RouteSynchronizer) getNexthopsMonopath(r *netlink.Route) (appldb.Nexthops, error) {
	ifaName, err := rr.ifNameResolver.ifNameByIndex(r.LinkIndex)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get interface by index")
	}

	return appldb.Nexthops{
		{
			Nexthop: r.Gw,
			IfName:  ifaName,
		},
	}, nil
}

func (rr *RouteSynchronizer) getNexthopsMultipath(r *netlink.Route) (appldb.Nexthops, error) {
	nexthops := make(appldb.Nexthops, len(r.MultiPath))

	for i := 0; i < len(r.MultiPath); i++ {
		ifaName, err := rr.ifNameResolver.ifNameByIndex(r.LinkIndex)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get interface by index")
		}

		nexthops[i].IfName = ifaName
		nexthops[i].Nexthop = r.MultiPath[i].Gw
	}

	return nexthops, nil
}
