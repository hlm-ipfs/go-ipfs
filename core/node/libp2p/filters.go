package libp2p

import (
	"strings"

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// filtersConnectionGater is an adapter that turns multiaddr.Filter into a
// connmgr.ConnectionGater.
type filtersConnectionGater ma.Filters

var _ connmgr.ConnectionGater = (*filtersConnectionGater)(nil)

func (f *filtersConnectionGater) InterceptAddrDial(p peer.ID, addr ma.Multiaddr) (allow bool) {
	rv1 := strings.HasPrefix(p.String(), "Qm")                                        //v1中继
	if rv1 && len(natInfo.Type) > 0 && !strings.Contains(natInfo.Type, "Symmetric") { //对称nat才连接v1中继
		return false
	}

	return !(*ma.Filters)(f).AddrBlocked(addr) && addr.Protocols()[0].Code == ma.P_IP4
}

func (f *filtersConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return true
}

func (f *filtersConnectionGater) InterceptAccept(connAddr network.ConnMultiaddrs) (allow bool) {
	return !(*ma.Filters)(f).AddrBlocked(connAddr.RemoteMultiaddr())
}

func (f *filtersConnectionGater) InterceptSecured(_ network.Direction, _ peer.ID, connAddr network.ConnMultiaddrs) (allow bool) {
	return !(*ma.Filters)(f).AddrBlocked(connAddr.RemoteMultiaddr())
}

func (f *filtersConnectionGater) InterceptUpgraded(_ network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}
