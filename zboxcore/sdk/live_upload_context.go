package sdk

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalContext listen syscall signal to cancel context
type SignalContext struct {
	context.Context
}

// NewSignalContext create SignalContext instance
func NewSignalContext(ctx context.Context) context.Context {

	sc := &SignalContext{}

	c, cancel := context.WithCancel(ctx)

	sc.Context = c

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-s
		cancel()
	}()

	return sc
}
