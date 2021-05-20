package appldb

import (
	"net"
	"strings"
)

// Nexthop represents a nexthop in the APPL_DB ROUTE object
type Nexthop struct {
	Nexthop net.IP
	IfName  string
}

// Nexthops represents a set of Nexthop
type Nexthops []Nexthop

// IfNames gets the space separated list of interface names
func (n *Nexthops) IfNames() string {
	ifNames := make([]string, len(*n))

	for i := 0; i < len(*n); i++ {
		ifNames[i] = (*n)[i].IfName
	}

	return strings.Join(ifNames, ",")
}

// Nexthops gets the space separated list of nexthops
func (n *Nexthops) Nexthops() string {
	nexthops := make([]string, len(*n))

	for i := 0; i < len(*n); i++ {
		nexthops[i] = (*n)[i].Nexthop.String()
	}

	return strings.Join(nexthops, ",")
}
