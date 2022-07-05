package libp2p

import (
	"context"
	logging "github.com/ipfs/go-log"
	manet "github.com/multiformats/go-multiaddr/net"
	"time"

	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)
var relaylog = logging.Logger("libp2p/relay")
const (
	relayTopic = "relay-discovery"
)

func RelayDiscovery(host host.Host, ps *pubsub.PubSub, peerChan AddrInfoChan) error {
	topic, err := ps.Join(relayTopic)
	if err != nil {
		return err
	}

	if sub, err := topic.Subscribe(); err != nil {
		return err
	} else {
		go relaySubLoop(sub, peerChan)
		go relayPubLoop(host, topic)
	}

	return nil
}

func relayPubLoop(host host.Host, topic *pubsub.Topic) {
	subReachability, _ := host.EventBus().Subscribe(new(event.EvtLocalReachabilityChanged))
	defer subReachability.Close()

	for {
		select {
		case ev, ok := <-subReachability.Out():
			if !ok {
				return
			}
			switch r := ev.(event.EvtLocalReachabilityChanged).Reachability; r {
			case network.ReachabilityPublic:
				info := peer.AddrInfo{
					ID:    host.ID(),
					Addrs: host.Addrs(),
				}
				addrs,err:=peer.AddrInfoToP2pAddrs(&info)
				if err!=nil{
					relaylog.Infow("relay pub msg ", "err", err)
					continue
				}
				for _,addr:=range addrs{
					if manet.IsPrivateAddr(addr){
						continue
					}
					msg:=addr.String()
					err := topic.Publish(context.TODO(), []byte(msg))
					relaylog.Infow("relay pub msg", "msg", msg, "err", err)
				}
			}
		}
	}
}

func relaySubLoop(sub *pubsub.Subscription, peerChan AddrInfoChan) {
	for range time.Tick(time.Second * 15) {
		msg, err := sub.Next(context.TODO())
		if err != nil {
			relaylog.Errorw("relay sub error", "err", err.Error())
		} else {
			relaylog.Infow("relay sub msg", "msg", string(msg.Data))
			if info, err := peerAddrInfo(string(msg.Data)); err != nil {
				relaylog.Errorw("relay sub addr-info invalid", "msg", string(msg.Data), "err", err.Error())
			} else {
				peerChan <- *info
			}
		}
	}
}

func peerAddrInfo(addrStr string) (*peer.AddrInfo, error) {
	var (
		err      error
		addr     multiaddr.Multiaddr
		addrInfo *peer.AddrInfo
	)
	if addr, err = multiaddr.NewMultiaddr(addrStr); err != nil {
		return nil, err
	}
	if addrInfo, err = peer.AddrInfoFromP2pAddr(addr); err != nil {
		return nil, err
	}
	return addrInfo, nil
}
