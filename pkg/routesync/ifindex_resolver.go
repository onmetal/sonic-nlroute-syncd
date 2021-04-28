package routesync

import (
	"net"

	"github.com/pkg/errors"
)

type ifNameResolver interface {
	ifNameByIndex(linkIndex int) (string, error)
}

type ifNameResolverNetlink struct {
}

func (inrn *ifNameResolverNetlink) ifNameByIndex(linkIndex int) (string, error) {
	ifa, err := net.InterfaceByIndex(linkIndex)
	if err != nil {
		return "", errors.Wrap(err, "Unable to get interface by index")
	}

	return ifa.Name, nil
}
