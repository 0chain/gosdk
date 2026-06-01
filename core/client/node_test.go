package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Regression: Fail()/Success() with an id that was not in the original holder
// (e.g. a sharder URL that differs from the configured one after a host/scheme
// rewrite) used to add the id to h.nodes WITHOUT an h.stats entry, so the next
// adjustNode dereferenced a nil *NodeStruct and panicked — fatal for any
// process using this SDK (blobber, validator, wasm: the whole Go runtime
// exits, surfacing as 502s and consensus_failed on every read).
func TestNodeHolder_AdjustUnregisteredIDNoPanic(t *testing.T) {
	h := NewHolder([]string{"1", "2", "3"}, 2)

	// The second of these used to panic before the fix.
	h.Fail("https://new-sharder.example/sharder01")
	h.Success("https://another-sharder.example/sharder01")
	h.Fail("1")

	// Every id in the ordered list must have a stats entry, else a future
	// adjustNode crashes.
	for _, id := range h.All() {
		assert.NotNilf(t, h.stats[id], "node %q in h.nodes has no h.stats entry", id)
	}
}
