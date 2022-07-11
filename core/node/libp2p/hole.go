package libp2p

import (
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	microLogger "go-micro.dev/v4/logger"

	"hlm-ipfs/x/infras"
	"hlm-ipfs/x/logger"
)

func NewHoleTrace() *HoleTrace {
	l, err := logger.Create(logger.LoggerKindZap, logger.LoggerFormatterJson, microLogger.TraceLevel, "logs", "hole", false)
	infras.Throw(err)

	return &HoleTrace{log: l}
}

type HoleTrace struct {
	log *logger.Logger
}

func (t *HoleTrace) Trace(evt *holepunch.Event) {
	t.log.Infow("hole punch trace", "event", evt)
}
