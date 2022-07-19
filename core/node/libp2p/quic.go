package libp2p

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	qlogging "github.com/lucas-clemente/quic-go/logging"
)

func NewQuicTrace() *QuicTrace {
	return &QuicTrace{log: logging.Logger("p2p-holepunch")}
}

type QuicTrace struct {
	log *logging.ZapEventLogger

	role   string
	connID string
}

func (t *QuicTrace) Trace(role qlogging.Perspective, connID []byte) io.WriteCloser {
	t.role = role.String()
	t.connID = fmt.Sprintf("%x", connID)
	return t
}

func (t *QuicTrace) Write(p []byte) (n int, err error) {
	str := strings.TrimSpace(string(p))
	if len(str) > 0 && str != "\n" {
		dict := make(map[string]interface{})
		if e := json.Unmarshal(p, &dict); e == nil {
			items := []interface{}{
				"role",
				t.role,
				"conn-id",
				t.connID,
			}
			for k, v := range dict {
				items = append(items, k, v)
			}
			t.log.Infow("quic info", items...)
		} else {
			t.log.Infow(str, "role", t.role, "conn-id", t.connID)
		}
	}
	return len(p), nil
}

func (t *QuicTrace) Close() error {
	return nil
}
