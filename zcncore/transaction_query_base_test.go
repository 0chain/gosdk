package zcncore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

const (
	keyServerAddr = "serverAddr"
	addrPrefix    = "http://localhost"
)

var tq *TransactionQuery
var numSharders int
var avgTimeToFindSharder float32
var maxTimePerIteration float32
var sharders []string

type SharderHealthStatus struct {
	Host         string `json:"host"`
	HealthStatus string `json:"health"`
}

func Setup(t *testing.T) {
	numSharders = 4
	var sharderPorts []string
	for i := 0; i < numSharders; i++ {
		port := fmt.Sprintf(":600%d", i)
		sharderPorts = append(sharderPorts, port)
		sharders = append(sharders, addrPrefix+port)
	}
	startMockSharderServers(sharders)
	// wait for 2s for all servers to start
	time.Sleep(2 * time.Second)
	var err error
	tq, err = NewTransactionQuery(sharders)
	if err != nil {
		t.Fatalf("Failed to create new transaction query: %v", err)
	}
}

func TestGetRandomSharder(t *testing.T) {
	Setup(t)
	for _, tc := range []struct {
		name            string
		offlineSharders []string
		expectErr       bool
		error           error
	}{
		{
			name:            "all sharders online",
			offlineSharders: []string{},
			expectErr:       false,
		},
		{
			name:            "only one sharder online",
			offlineSharders: []string{"http://localhost:6000", "http://localhost:6002", "http://localhost:6003"},
			expectErr:       false,
		},
		{
			name:            "all sharders offline",
			offlineSharders: sharders,
			expectErr:       true,
			error:           ErrNoOnlineSharders,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, host := range tc.offlineSharders {
				tq.offline[host] = true
			}
			var onlineSharders []string
			for _, s := range sharders {
				if !contains(tc.offlineSharders, s) {
					onlineSharders = append(onlineSharders, s)
				}
			}
			sharder, err := tq.getRandomSharder(context.Background())
			if !tc.expectErr {
				require.NoError(t, err)
				require.Subset(t, onlineSharders, []string{sharder})
			} else {
				require.EqualError(t, err, tc.error.Error())
			}
		})
	}
}

func TestGetRandomSharderAndBenchmark(t *testing.T) {
	Setup(t)

	done := make(chan struct{})
	go startAndStopShardersRandomly(done)
	go waitSignal(done)
	fetchRandomSharderAndBenchmark(t)
	<-done
}

func startMockSharderServers(sharders []string) {
	for i, _ := range sharders {
		url := fmt.Sprintf(":600%d", i)
		go func(url string) {
			ctx, cancel := context.WithCancel(context.Background())
			mx := http.NewServeMux()
			mx.HandleFunc(SharderEndpointHealthCheck, getSharderHealth)
			httpServer := &http.Server{
				Addr:    url,
				Handler: mx,
				BaseContext: func(l net.Listener) context.Context {
					ctx := context.WithValue(ctx, keyServerAddr, url)
					return ctx
				},
			}
			log.Printf("Starting server at: %v", url)
			err := httpServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				log.Printf("server one closed\n")
			} else if err != nil {
				log.Printf("error listening for server one: %s\n", err)
			}
			cancel()
		}(url)
	}
}

func getSharderHealth(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sharderHost := ctx.Value(keyServerAddr).(string)
	tq.RLock()
	_, ok := tq.offline[sharderHost]
	tq.RUnlock()
	if ok {
		errorAny(w, 404, fmt.Sprintf("sharder %v is offline", sharderHost))
	} else {
		healthStatus := &SharderHealthStatus{
			Host:         sharderHost,
			HealthStatus: "healthy",
		}
		err := json.NewEncoder(w).Encode(healthStatus)
		if err != nil {
			errorAny(w, http.StatusInternalServerError, "failed to encode json")
		}
	}
}

func startAndStopShardersRandomly(done chan struct{}) {
	for {
		select {
		case <-time.After(5 * time.Millisecond):
			tq.Lock()
			// mark a random sharder offline every 5ms
			randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
			randomSharder := tq.sharders[randGen.Intn(numSharders)]
			tq.offline[randomSharder] = true
			tq.Unlock()

		case <-time.After(3 * time.Millisecond):
			tq.Lock()
			// mark a random sharder online every 3ms
			randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
			randomSharder := tq.sharders[randGen.Intn(numSharders)]
			if _, ok := tq.offline[randomSharder]; ok {
				delete(tq.offline, randomSharder)
			}
			tq.Unlock()

		case <-time.After(5 * time.Second):
			//Randomly mark all sharders online every 5s
			tq.Lock()
			tq.Reset()
			tq.Unlock()
		case <-done:
			return
		}
	}
}

func fetchRandomSharderAndBenchmark(t *testing.T) {
	numIterations := 10
	for i := 0; i < numIterations; i++ {
		// Sleep for sometime to have some random sharders started and stopped
		time.Sleep(20 * time.Millisecond)
		ctx := context.Background()
		start := time.Now()
		_, err := tq.getRandomSharder(ctx)
		if err != nil {
			t.Fatalf("Failed to get a random sharder err: %v", err)
		}
		end := float32(time.Since(start) / time.Microsecond)
		if end > maxTimePerIteration {
			maxTimePerIteration = end
		}
		avgTimeToFindSharder += end

	}
	avgTimeToFindSharder = (avgTimeToFindSharder / float32(numIterations)) / 1000
	maxTimePerIteration /= 1000
	t.Logf("Average time to find a random sharder: %vms and max time for an iteration: %vms", avgTimeToFindSharder, maxTimePerIteration)
}

func errorAny(w http.ResponseWriter, status int, msg string) {
	httpMsg := fmt.Sprintf("%d %s", status, http.StatusText(status))
	if msg != "" {
		httpMsg = fmt.Sprintf("%s - %s", httpMsg, msg)
	}
	http.Error(w, httpMsg, status)
}

func waitSignal(stop chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	close(stop)
}

func contains(list []string, e string) bool {
	for _, l := range list {
		if l == e {
			return true
		}
	}
	return false
}
