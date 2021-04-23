package main

import (
	"fmt"
	"net"
	"syscall"

	"github.com/go-redis/redis"
	"github.com/vishvananda/netlink"

	log "github.com/sirupsen/logrus"
)

func main() {
	routesCh := make(chan netlink.RouteUpdate)

	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	err := netlink.RouteSubscribe(routesCh, nil)
	if err != nil {
		log.WithError(err).Panic("Unable to subscribe to netlink route updates")
	}

	for {
		u := <-routesCh

		fmt.Printf("Update type: %d\n", u.Type)
		fmt.Printf("Route: %v\n", u)

		if u.Type == syscall.RTM_NEWROUTE {
			addRoute(rc, u.Route)
			continue
		}

		if u.Type == syscall.RTM_DELROUTE {
			delRoute(rc, u.Route)
			continue
		}

	}
}

func addRoute(rc *redis.Client, u netlink.Route) {
	pfxStr := u.Dst.String()
	ifname, err := net.InterfaceByIndex(u.LinkIndex)
	if err != nil {
		log.WithError(err).Error("Unable to get interface name")
	}

	err = rc.SAdd("ROUTE_TABLE_KEY_SET", pfxStr).Err()
	if err != nil {
		log.WithError(err).Error("Redis SADD call failed")
		return
	}

	err = rc.HSet(fmt.Sprintf("_ROUTE_TABLE:%s", pfxStr), "nexthop", u.Gw.String()).Err()
	if err != nil {
		log.WithError(err).Error("Redis HSET call failed")
		return
	}

	err = rc.HSet(fmt.Sprintf("_ROUTE_TABLE:%s", pfxStr), "ifname", ifname.Name).Err()
	if err != nil {
		log.WithError(err).Error("Redis HSET call failed")
		return
	}

	err = rc.Publish("ROUTE_TABLE_CHANNEL", "G").Err()
	if err != nil {
		log.WithError(err).Error("Redis PUBLISH call failed")
		return
	}
}

func delRoute(rc *redis.Client, u netlink.Route) {
	pfxStr := u.Dst.String()

	err := rc.SAdd("ROUTE_TABLE_KEY_SET", pfxStr).Err()
	if err != nil {
		log.WithError(err).Error("Redis SADD call failed")
		return
	}

	err = rc.SAdd("ROUTE_TABLE_DEL_SET", pfxStr).Err()
	if err != nil {
		log.WithError(err).Error("Redis SADD call failed")
		return
	}

	err = rc.Del(fmt.Sprintf("_ROUTE_TABLE:%s", pfxStr)).Err()
	if err != nil {
		log.WithError(err).Error("Redis DEL call failed")
		return
	}

	err = rc.Publish("ROUTE_TABLE_CHANNEL", "G").Err()
	if err != nil {
		log.WithError(err).Error("Redis PUBLISH call failed")
		return
	}
}
