package appldb

import (
	"fmt"
	"net"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type APPLDB struct {
	rc *redis.Client
}

// New creates a new APPL_DB handler
func New() *APPLDB {
	return &APPLDB{
		rc: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
	}
}

func (a *APPLDB) hsetMap(key string, kv map[string]string) error {
	var err error

	for k, v := range kv {
		err = a.rc.HSet(key, k, v).Err()
		if err != nil {
			return errors.Wrap(err, "HSET call failed")
		}
	}

	return nil
}

// AddRoute adds a route
func (a *APPLDB) AddRoute(pfx net.IPNet, nexthops Nexthops) error {
	pfxStr := pfx.String()

	err := a.rc.SAdd("ROUTE_TABLE_KEY_SET", pfxStr).Err()
	if err != nil {
		return errors.Wrap(err, "SADD call failed")
	}

	key := fmt.Sprintf("_ROUTE_TABLE:%s", pfxStr)

	err = a.hsetMap(key, map[string]string{
		"nexthop": nexthops.Nexthops(),
		"ifname":  nexthops.IfNames(),
	})
	if err != nil {
		return errors.Wrap(err, "hsetMap failed")
	}

	err = a.rc.Publish("ROUTE_TABLE_CHANNEL", "G").Err()
	if err != nil {
		return errors.Wrap(err, "PUBLISH call failed")
	}

	return nil
}

func (a *APPLDB) DelRoute(pfx net.IPNet) error {
	pfxStr := pfx.String()

	err := a.rc.SAdd("ROUTE_TABLE_KEY_SET", pfxStr).Err()
	if err != nil {
		return errors.Wrap(err, "SADD failed")
	}

	err = a.rc.SAdd("ROUTE_TABLE_DEL_SET", pfxStr).Err()
	if err != nil {
		return errors.Wrap(err, "SADD failed")
	}

	err = a.rc.Del(fmt.Sprintf("_ROUTE_TABLE:%s", pfxStr)).Err()
	if err != nil {
		return errors.Wrap(err, "DEL failed")
	}

	err = a.rc.Publish("ROUTE_TABLE_CHANNEL", "G").Err()
	if err != nil {
		return errors.Wrap(err, "PUBLISH failed")
	}

	return nil
}
