package libp2p

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/control"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	ConnGater = simpleOpt(DefaultConnGater)
)

type MyConnGater struct {
}

func (g *MyConnGater) InterceptPeerDial(peer.ID) bool {
	return true
}

func (g *MyConnGater) InterceptAddrDial(p peer.ID, addr ma.Multiaddr) bool {
	return addr.Protocols()[0].Code == ma.P_IP4
}

func (g *MyConnGater) InterceptAccept(network.ConnMultiaddrs) bool {
	return true
}

func (g *MyConnGater) InterceptSecured(network.Direction, peer.ID, network.ConnMultiaddrs) bool {
	return true
}

func (g *MyConnGater) InterceptUpgraded(network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}

func DefaultConnGater(cfg *libp2p.Config) error {
	opt := libp2p.ConnectionGater(&MyConnGater{})
	return cfg.Apply(opt)
}
