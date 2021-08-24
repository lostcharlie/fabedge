package routing

import (
	"github.com/fabedge/fabedge/pkg/tunnel"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
)

type Routing interface {
	SyncRoutes(active bool, connections []tunnel.ConnConfig) error
	CleanRoutes(connections []tunnel.ConnConfig) error
}

func IsInConns(dst *net.IPNet, connections []tunnel.ConnConfig) (bool, error) {
	for _, con := range connections {
		for _, subnet := range con.RemoteSubnets {
			s, err := netlink.ParseIPNet(subnet)
			if err != nil {
				return false, err
			}
			if s.String() == dst.String() {
				return true, nil
			}
		}
	}
	return false, nil
}

func addAllEdgeRoutes(conns []tunnel.ConnConfig, table int) error {
	gw, err := getDefaultGateway()
	if err != nil {
		return err
	}

	for _, conn := range conns {
		for _, subnet := range conn.RemoteSubnets {
			s, err := netlink.ParseIPNet(subnet)
			if err != nil {
				return err
			}
			// add into table 220
			route := netlink.Route{Dst: s, Gw: gw, Table: table}
			err = netlink.RouteAdd(&route)
			if err != nil && !fileExistsError(err) {
				return err
			}
		}
	}

	return nil
}

func EnsureDummyDevice(devName string) error {
	link, err := netlink.LinkByName(devName)
	if err == nil {
		return netlink.LinkSetUp(link)
	}

	link = &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{Name: devName},
	}
	if err = netlink.LinkAdd(link); err != nil {
		return err
	}

	return netlink.LinkSetUp(link)
}

func delEdgeRoute(subnet *net.IPNet) error {
	gw, err := getDefaultGateway()
	if err != nil {
		return err
	}
	route := netlink.Route{Dst: subnet, Gw: gw, Table: TableStrongswan}
	return netlink.RouteDel(&route)
}

func delAllEdgeRoutes(conns []tunnel.ConnConfig) error {
	for _, conn := range conns {
		for _, subnet := range conn.RemoteSubnets {
			s, err := netlink.ParseIPNet(subnet)
			if err != nil {
				return err
			}
			err = delEdgeRoute(s)
			if err != nil && !noSuchProcessError(err) {
				return err
			}
		}
	}

	return nil
}

func getDefaultGateway() (net.IP, error) {
	defaultRoute, err := netlink.RouteGet(net.ParseIP("8.8.8.8"))
	if len(defaultRoute) != 1 || err != nil {
		return nil, err
	}
	return defaultRoute[0].Gw, nil
}

func fileExistsError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "file exists")
}

// occur when the route does not exist
func noSuchProcessError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "no such process")
}

// occur when the link does not exist
func invalidArgument(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "invalid argument")
}

func delRoutesNotInConnections(connections []tunnel.ConnConfig, table int) error {
	var routeFilter = &netlink.Route{
		Table: table,
	}
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, routeFilter, netlink.RT_FILTER_TABLE)
	if err != nil {
		return err
	}

	for _, r := range routes {
		if yes, err := IsInConns(r.Dst, connections); err == nil && !yes {
			err = delEdgeRoute(r.Dst)
		}
	}

	return err
}
