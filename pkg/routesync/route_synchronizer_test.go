package routesync

import (
	"fmt"
	"net"
	"testing"

	"github.com/onmetal/sonic-nlroute-syncd/pkg/appldb"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

type mockIfNameResolver struct {
}

func (minr *mockIfNameResolver) ifNameByIndex(linkIndex int) (string, error) {
	switch linkIndex {
	case 0:
		return "lo", nil
	case 1:
		return "eth0", nil
	default:
		return "", fmt.Errorf("Unable to resolve interface index")
	}
}

func TestGetNexthops(t *testing.T) {
	tests := []struct {
		name     string
		input    *netlink.Route
		expected appldb.Nexthops
		wantFail bool
	}{
		{
			name: "Test #1",
			input: &netlink.Route{
				Gw:        net.IP{192, 0, 2, 0},
				LinkIndex: 0,
			},
			expected: appldb.Nexthops{
				{
					Nexthop: net.IP{192, 0, 2, 0},
					IfName:  "lo",
				},
			},
			wantFail: false,
		},
		{
			name: "Test #2",
			input: &netlink.Route{
				Gw:        net.IP{192, 0, 2, 1},
				LinkIndex: 1,
			},
			expected: appldb.Nexthops{
				{
					Nexthop: net.IP{192, 0, 2, 1},
					IfName:  "eth0",
				},
			},
			wantFail: false,
		},
		{
			name: "Test #3",
			input: &netlink.Route{
				Gw:        net.IP{192, 0, 2, 2},
				LinkIndex: 2,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		rs := New(nil)
		rs.ifNameResolver = &mockIfNameResolver{}

		res, err := rs.getNexthops(test.input)
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q", test.name)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expected, res, test.name)
	}
}
