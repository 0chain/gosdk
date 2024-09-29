// DEPRECATED: This package is deprecated and will be removed in a future release.
package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/0chain/gosdk/zmagmacore/log"
)

type (
	// Closable represents interface for types that might be closed.
	Closable interface {
		Close() error
	}
)

// Handle handles various shutdown signals.
func Handle(ctx context.Context, server *http.Server, grpcServer *grpc.Server, closable ...Closable) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// wait for signals or app context done
	select {
	case <-ctx.Done():
	case <-c:
	}
	shutdown(server, grpcServer, closable...)
}

func shutdown(server *http.Server, grpcServer *grpc.Server, closable ...Closable) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Logger.Warn("Server failed to gracefully shuts down", zap.Error(err))
	}
	log.Logger.Debug("Server is shut down.")

	grpcServer.GracefulStop()
	log.Logger.Debug("GRPC server is shut down.")

	log.Logger.Debug("Closing rest ...")
	for _, cl := range closable {
		if err := cl.Close(); err != nil {
			log.Logger.Warn("Can not close.", zap.String("err", err.Error()))
		}
	}
}
