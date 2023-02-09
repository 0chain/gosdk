package zcncore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type SharderStatus struct {
	SharderHost  string `json:"host"`
	HealthStatus string `json:"health"`
}

func TestGettingRandomSharder(t *testing.T) {
	numSharders = 20
	var sharderPorts []string
	var sharders []string
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

	done := make(chan struct{})
	go startAndStopShardersRandomly(t, done)
	go fetchRandomSharderAndBenchmark(t, done)
	go waitSignal(done)
	<-done
}

func startMockSharderServers(sharders []string) {
	for i, _ := range sharders {
		url := fmt.Sprintf(":600%d", i)
		go func(url string) {
			ctx, cancel := context.WithCancel(context.Background())
			mx := http.NewServeMux()
			mx.HandleFunc(SharderEndpointHealthCheck, returnSharderHealth)
			httpServer := &http.Server{
				Addr:    url,
				Handler: mx,
				BaseContext: func(l net.Listener) context.Context {
					ctx := context.WithValue(ctx, keyServerAddr, url)
					return ctx
				},
			}
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

func returnSharderHealth(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sharderHost := ctx.Value(keyServerAddr).(string)
	tq.RLock()
	_, ok := tq.offline[sharderHost]
	tq.RUnlock()
	if ok {
		errorAny(w, 404, fmt.Sprintf("sharder %v is offline", sharderHost))
	} else {
		healthStatus := &SharderStatus{
			SharderHost:  sharderHost,
			HealthStatus: "healthy",
		}
		err := json.NewEncoder(w).Encode(healthStatus)
		if err != nil {
			errorAny(w, http.StatusInternalServerError, "failed to encode json")
		}
	}
}

func startAndStopShardersRandomly(t *testing.T, done chan struct{}) {
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
			//Randomly mark all sharders online after every 5 seconds
			t.Logf("Marking all sharders online")
			tq.Lock()
			tq.Reset()
			tq.Unlock()
		case <-done:
			return
		}
	}
}

func fetchRandomSharderAndBenchmark(t *testing.T, done chan struct{}) {
	numIterations := 500
	for i := 0; i < numIterations; i++ {
		select {
		case <-time.After(20 * time.Millisecond):
			ctx := context.Background()
			start := time.Now()
			randomSharder, err := tq.randOne(ctx)
			if err != nil {
				t.Fatalf("Failed to get a random sharder err: %v", err)
			}
			end := time.Since(start)
			t.Logf("Found sharder %v online. Time taken for %v sharders: %v", randomSharder, numSharders, end)
			avgTimeToFindSharder += float32(end / time.Microsecond)
		}
	}
	avgTimeToFindSharder = (avgTimeToFindSharder / float32(numIterations)) / 1000
	t.Logf("Average time to find sharder: %vms", avgTimeToFindSharder)
	close(done)
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
