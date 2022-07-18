package libp2p

import (
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	microLogger "go-micro.dev/v4/logger"
	"os"
	"path/filepath"

	"hlm-ipfs/x/infras"
	"hlm-ipfs/x/logger"
)

func NewHoleTrace() *HoleTrace {
	dir := "logs"
	if str, ok := os.LookupEnv("IPFS_PATH"); ok && len(str) > 0 {
		dir = filepath.Join(str, "log")
		_ = os.Mkdir(dir, os.ModePerm)
	}
	l, err := logger.Create(logger.LoggerKindZap, logger.LoggerFormatterJson, microLogger.TraceLevel, dir, "hole", false)
	infras.Throw(err)

	return &HoleTrace{log: l}
}

type HoleTrace struct {
	log *logger.Logger
}

func (t *HoleTrace) Trace(evt *holepunch.Event) {
	t.log.Infow("hole punch trace", "event", evt)
}
