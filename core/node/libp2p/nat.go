package libp2p

import (
	"time"

	"github.com/ccding/go-stun/stun"
	config "github.com/ipfs/kubo/config"
	"github.com/libp2p/go-libp2p"
)

var NatPortMap = simpleOpt(libp2p.NATPortMap())

func AutoNATService(throttle *config.AutoNATThrottleConfig) func() Libp2pOpts {
	return func() (opts Libp2pOpts) {
		opts.Opts = append(opts.Opts, libp2p.EnableNATService())
		if throttle != nil {
			opts.Opts = append(opts.Opts,
				libp2p.AutoNATServiceRateLimit(
					throttle.GlobalLimit,
					throttle.PeerLimit,
					throttle.Interval.WithDefault(time.Minute),
				),
			)
		}
		return opts
	}
}

type NatInfo struct {
	Type string
	Addr string
}

var (
	natInfo = &NatInfo{}
)

func init() {
	/*if info := CheckNat(); info != nil {
		natInfo = info
	}*/
}

func CheckNat() *NatInfo {
	c := stun.NewClient()
	c.SetServerAddr("stun.stunprotocol.org:3478")
	nat, host, err := c.Discover()
	if err != nil || host == nil {
		return nil
	}
	return &NatInfo{
		Type: nat.String(),
		Addr: host.String(),
	}
}
