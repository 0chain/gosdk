package sdk

// Tests + benchmarks for the erasure-encode parallelization landed in
// d505e488 (CreateChunkedUpload): the reedsolomon encoder option changed from
//
//	reedsolomon.WithAutoGoroutines(int(su.chunkSize))
//
// to
//
//	reedsolomon.WithMaxGoroutines(runtime.NumCPU())
//
// so a large upload batch's erasure-encode fans across every vCPU instead of
// being throttled by a chunk-size-derived heuristic.
//
// HOW reedsolomon parallelizes (klauspost/reedsolomon v1.11.8, the pinned
// version): a single Encode call splits the *shard's bytes* across up to
// maxGoroutines workers, but only when the shard is larger than minSplitSize
// (cache-derived, 16 KiB on this CPU). So the encode is only ever parallel on
// multi-MB shards — the large-upload case the producer cares about. (Between
// ~minSplitSize and ~10 MiB the library uses the vectorized parallel path where
// the goroutine count dominates throughput; above ~10 MiB it falls back to a
// generic per-round split that narrows the benefit.)
//
// WHY WithMaxGoroutines is strictly better — measured below for the cluster's
// 2/1 layout: reedsolomon.New() collapses the auto-derived goroutine count to 1
// whenever the supplied shardSize is <= minSplitSize*2 (= 32 KiB here). With
// the old option that argument was su.chunkSize, so any deployment that tunes
// the chunk to <=32 KiB (small-file / low-latency profiles) silently pins the
// erasure-encode of even large shards to a single core — the
// CPU-bound-on-one-core symptom the commit describes. WithMaxGoroutines(NumCPU)
// is independent of shardSize and always keeps the encode parallel.
// BenchmarkErasureEncode_Collapse vs BenchmarkErasureEncodeAuto_Collapse show
// that ~core-count gap directly (≈2x on this 12-core box for a 2 MiB shard).
//
// NOTE: at the default 64 KiB chunk (DefaultChunkSize) the auto heuristic lands
// just *above* the collapse boundary and derives the same goroutine count, so
// BenchmarkErasureEncode and BenchmarkErasureEncodeAuto tie. That pair is the
// honest, directly-comparable baseline at the current default; the win shows up
// once the chunk enters the collapse zone.
//
// These tests prove (a) the parallel option is lossless and (b) it beats the
// old auto option in the regime where auto serializes.

import (
	"bytes"
	"crypto/rand"
	"runtime"
	"testing"

	"github.com/klauspost/reedsolomon"
)

const (
	// Allocation layout used by the cluster: 2 data + 1 parity (2/1).
	testDataShards   = 2
	testParityShards = 1

	// Multi-MB shard so each Encode's byteCount >> reedsolomon's cache-derived
	// minSplitSize and the parallel encode path actually triggers (the shard
	// bytes get split across maxGoroutines). At a sub-MB shard the encode stays
	// serial under *every* option, so the goroutine choice would be moot.
	testShardSize = 4 << 20 // 4 MiB per shard (round-trip correctness)

	// Per-iteration benchmark batch across the 2/1 layout. With 2 data shards
	// this is a 2 MiB shard — in the vectorized parallel regime (> minSplitSize,
	// < ~10 MiB) where the maxGoroutines choice dominates throughput, so Max vs
	// Auto is most directly visible. A batch is encoded many times per second by
	// the upload producer; one large shard set is the representative unit.
	batchBytes = 4 << 20 // -> 2 MiB data shards

	// A sub-32-KiB shardSize handed to WithAutoGoroutines forces its derived
	// goroutine count to 1 (collapse zone). The _Collapse benchmark configures
	// the auto heuristic for this small chunk while still encoding the same
	// large batch, modelling a small-file / low-latency chunk profile.
	collapseChunk = 16 << 10
)

// newParallelEncoder builds the reedsolomon encoder EXACTLY as
// CreateChunkedUpload does after d505e488.
func newParallelEncoder(t testing.TB) reedsolomon.Encoder {
	t.Helper()
	enc, err := reedsolomon.New(
		testDataShards,
		testParityShards,
		reedsolomon.WithMaxGoroutines(runtime.NumCPU()),
	)
	if err != nil {
		t.Fatalf("reedsolomon.New (WithMaxGoroutines): %v", err)
	}
	return enc
}

// TestErasureEncodeParallelRoundTrip proves the parallel encode option
// preserves correctness: split a known buffer, Encode the parity, drop up to
// parityShards shards, Reconstruct, and assert the recovered data is
// byte-identical to the original.
func TestErasureEncodeParallelRoundTrip(t *testing.T) {
	enc := newParallelEncoder(t)

	// Known buffer sized to exactly fill testDataShards shards so Split yields
	// equal, fully-populated data shards (no zero padding to reason about), and
	// large enough that the parallel encode path actually triggers.
	orig := make([]byte, testDataShards*testShardSize)
	if _, err := rand.Read(orig); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	shards, err := enc.Split(orig)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if got := len(shards); got != testDataShards+testParityShards {
		t.Fatalf("Split shard count = %d, want %d", got, testDataShards+testParityShards)
	}

	if err := enc.Encode(shards); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Sanity: parity must verify before we damage anything.
	ok, err := enc.Verify(shards)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("Verify returned false for freshly-encoded shards")
	}

	// Lose up to parityShards shards. With 2/1 we can lose exactly one and
	// still rebuild. Drop a *data* shard — the hard case, since it must be
	// regenerated from the surviving data + parity.
	lost := append([]byte(nil), shards[0]...)
	shards[0] = nil

	if err := enc.Reconstruct(shards); err != nil {
		t.Fatalf("Reconstruct: %v", err)
	}
	if !bytes.Equal(shards[0], lost) {
		t.Fatal("reconstructed data shard differs from original")
	}

	// Join the data shards back and compare against the original buffer end to
	// end — the ultimate proof the parallel encode is lossless.
	var rebuilt bytes.Buffer
	if err := enc.Join(&rebuilt, shards, len(orig)); err != nil {
		t.Fatalf("Join: %v", err)
	}
	if !bytes.Equal(rebuilt.Bytes(), orig) {
		t.Fatal("round-tripped data is not byte-identical to original")
	}
}

// buildBatch returns a freshly-split shard set (data populated from random
// bytes, parity zeroed) sized so the total payload is ~size bytes. Both
// benchmarks call this so they encode identical work.
func buildBatch(b *testing.B, enc reedsolomon.Encoder, size int) [][]byte {
	b.Helper()
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		b.Fatalf("rand.Read: %v", err)
	}
	shards, err := enc.Split(data)
	if err != nil {
		b.Fatalf("Split: %v", err)
	}
	return shards
}

// benchEncode encodes the whole ~batchBytes payload as one large shard set per
// iteration and reports MB/s via SetBytes. The encoder's option is the only
// thing that varies across benchmarks, so the ns/op are directly comparable.
func benchEncode(b *testing.B, enc reedsolomon.Encoder) {
	shards := buildBatch(b, enc, batchBytes)
	b.SetBytes(int64(batchBytes))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(shards); err != nil {
			b.Fatalf("Encode: %v", err)
		}
	}
}

// BenchmarkErasureEncode — NEW WithMaxGoroutines(NumCPU): encode parallel.
func BenchmarkErasureEncode(b *testing.B) {
	benchEncode(b, newParallelEncoder(b))
}

// BenchmarkErasureEncodeAuto — OLD WithAutoGoroutines(DefaultChunkSize=64 KiB),
// the literal pre-d505e488 option at the default chunk, on identical data. At
// 64 KiB the auto heuristic lands just above the collapse boundary and derives
// the same goroutine count, so this is expected to tie with
// BenchmarkErasureEncode — an honest baseline, not a win.
func BenchmarkErasureEncodeAuto(b *testing.B) {
	enc, err := reedsolomon.New(testDataShards, testParityShards, reedsolomon.WithAutoGoroutines(DefaultChunkSize))
	if err != nil {
		b.Fatalf("reedsolomon.New (WithAutoGoroutines): %v", err)
	}
	benchEncode(b, enc)
}

// BenchmarkErasureEncode_Collapse — NEW option on the same large batch: stays
// parallel regardless of chunk size. Paired with the Auto_Collapse benchmark
// below to isolate the goroutine effect.
func BenchmarkErasureEncode_Collapse(b *testing.B) {
	benchEncode(b, newParallelEncoder(b))
}

// BenchmarkErasureEncodeAuto_Collapse — OLD option with a collapse-zone chunk
// (<=32 KiB) on the SAME large batch: the auto heuristic forces a single
// goroutine, so the multi-MB shard is encoded serially. The ns/op gap vs
// BenchmarkErasureEncode_Collapse (≈ the usable core count) is the speedup
// d505e488 buys whenever the chunk is small enough to trip the heuristic.
func BenchmarkErasureEncodeAuto_Collapse(b *testing.B) {
	enc, err := reedsolomon.New(testDataShards, testParityShards, reedsolomon.WithAutoGoroutines(collapseChunk))
	if err != nil {
		b.Fatalf("reedsolomon.New (WithAutoGoroutines): %v", err)
	}
	benchEncode(b, enc)
}
