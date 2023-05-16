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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

func TestMain(m *testing.M) {
	numSharders = 10
	sharders = make([]string, 0)
	for i := 0; i < numSharders; i++ {
		port := fmt.Sprintf(":600%d", i)
		sharders = append(sharders, addrPrefix+port)
	}
	startMockSharderServers(sharders)
	// wait for 2s for all servers to start
	time.Sleep(2 * time.Second)
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestGetRandomSharder(t *testing.T) {
	var err error
	tq, err = NewTransactionQuery(sharders, []string{})
	if err != nil {
		t.Fatalf("Failed to create new transaction query: %v", err)
	}

	for _, tc := range []struct {
		name           string
		onlineSharders []string
		expectedErr    error
		setupContext   func(ctx context.Context) context.Context
	}{
		{
			name:           "context deadline exceeded",
			onlineSharders: []string{"http://localhost:6009"},
			expectedErr:    context.DeadlineExceeded,
			setupContext: func(ct context.Context) context.Context {
				ctx, cancel := context.WithTimeout(ct, 100*time.Microsecond)
				go func() {
					<-ctx.Done()
					cancel()
				}()
				return ctx
			},
		},
		{
			name:           "all sharders online",
			onlineSharders: sharders,
			expectedErr:    nil,
		},
		{
			name:           "only one sharder online",
			onlineSharders: []string{"http://localhost:6000"},
			expectedErr:    nil,
		},
		{
			name:           "few sharders online",
			onlineSharders: []string{"http://localhost:6001", "http://localhost:6006", "http://localhost:6009"},
			expectedErr:    nil,
		},
		{
			name:           "all sharders offline",
			onlineSharders: []string{},
			expectedErr:    ErrNoOnlineSharders,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tq.Reset()

			for _, s := range sharders {
				if !contains(tc.onlineSharders, s) {
					tq.Lock()
					tq.offline[s] = true
					tq.Unlock()
				}
			}
			ctx := context.Background()
			if tc.setupContext != nil {
				ctx = tc.setupContext(ctx)
			}
			sharder, err := tq.getRandomSharder(ctx)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				require.Subset(t, tc.onlineSharders, []string{sharder})
			} else {
				require.EqualError(t, err, tc.expectedErr.Error())
			}
		})
	}
}

// Maybe replace this with the standard go benchmark later on
func TestGetRandomSharderAndBenchmark(t *testing.T) {
	var err error
	tq, err = NewTransactionQuery(sharders, []string{})
	if err != nil {
		t.Fatalf("Failed to create new transaction query: %v", err)
	}

	done := make(chan struct{})
	go startAndStopShardersRandomly(done)
	fetchRandomSharderAndBenchmark(t)
	close(done)
}

func startMockSharderServers(sharders []string) {
	for i := range sharders {
		url := fmt.Sprintf(":600%d", i)
		go func(url string) {
			ctx, cancel := context.WithCancel(context.Background())
			mx := http.NewServeMux()
			mx.HandleFunc(SharderEndpointHealthCheck, getSharderHealth)
			httpServer := &http.Server{
				Addr:    url,
				Handler: mx,
				BaseContext: func(l net.Listener) context.Context {
					ctx := context.WithValue(ctx, keyServerAddr, url) // nolint
					return ctx
				},
			}
			log.Printf("Starting sharder server at: %v", url)
			err := httpServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				log.Printf("server %v closed\n", httpServer.Addr)
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
			delete(tq.offline, randomSharder)
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
	numIterations := 5
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

func contains(list []string, e string) bool {
	for _, l := range list {
		if l == e {
			return true
		}
	}
	return false
}
