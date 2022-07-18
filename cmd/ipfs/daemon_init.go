package main

import (
	"os"
	"time"

	"hlm-ipfs/x/infras"

	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
)

func init() {
	identify.ActivationThresh = 1

	holepunch.MaxRetries = 3
	holepunch.DialTimeout = time.Second * 6
	libp2pquic.HolePunchTimeout = time.Second * 6
	libp2pquic.QuicConfig.HandshakeIdleTimeout = time.Second * 6

	key := "QUIC_AESECB_KEY"
	if str, ok := os.LookupEnv(key); !ok || len(str) == 0 {
		infras.Throw(os.Setenv(key, "album_unwind_fret"))
	}
}
