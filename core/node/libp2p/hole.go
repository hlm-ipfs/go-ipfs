package libp2p

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
)

func NewHoleTrace() *HoleTrace {
	return &HoleTrace{log: logging.Logger("p2p-holepunch")}
}

type HoleTrace struct {
	log *logging.ZapEventLogger
}

func (t *HoleTrace) Trace(evt *holepunch.Event) {
	t.log.Infow("hole punch trace", "event", evt)
}
