package http

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/0chain/gosdk/zmagmacore/config"
	"github.com/0chain/gosdk/zmagmacore/log"
)

type setupHandlers func(r *mux.Router, cfg config.Handler)

// CreateServer creates http.Server and setups handlers.
func CreateServer(setupHandlers setupHandlers, cfg config.Handler, port int, development bool) *http.Server {
	// setup CORS
	router := mux.NewRouter()
	setupHandlers(router, cfg)

	address := ":" + strconv.Itoa(port)
	originsOk := handlers.AllowedOriginValidator(isValidOrigin)
	headersOk := handlers.AllowedHeaders([]string{
		"X-Requested-With", "X-App-Client-ID",
		"X-App-Client-Key", "Content-Type",
	})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	server := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20,
		Handler:           handlers.CORS(originsOk, headersOk, methodsOk)(router),
	}
	if development { // non idle & write timeouts setup to enable pprof
		server.IdleTimeout = 0
		server.WriteTimeout = 0
	}

	log.Logger.Info("Ready to listen to the requests")

	return server
}

// StartServer calls http.Server.ListenAndServe and calls app context cancel if error occurs.
func StartServer(server *http.Server, appCtxCancel func()) {
	err := server.ListenAndServe()
	if err != nil {
		log.Logger.Warn(err.Error())
		appCtxCancel()
	}
}

func isValidOrigin(origin string) bool {
	uri, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := uri.Hostname()
	switch { // allowed origins
	case host == "localhost":
	case host == "0chain.net":
	case strings.HasSuffix(host, ".0chain.net"):
	case strings.HasSuffix(host, ".alphanet-0chain.net"):
	case strings.HasSuffix(host, ".devnet-0chain.net"):
	case strings.HasSuffix(host, ".testnet-0chain.net"):
	case strings.HasSuffix(host, ".mainnet-0chain.net"):

	default: // not allowed
		return false
	}

	return true
}
