package helpers

import (
	"context"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"go.uber.org/fx"
)

type MetricsCtx context.Context

// LifecycleCtx creates a context which will be cancelled when lifecycle stops
//
// This is a hack which we need because most of our services use contexts in a
// wrong way
func LifecycleCtx(mctx MetricsCtx, lc fx.Lifecycle) context.Context {
	ctx, cancel := context.WithCancel(mctx)
	//ctx = network.WithUseTransient(ctx, "test")
	ctx = swarm.WithoutV2Relay(ctx)
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
	return ctx
}
