package zbox

import (
	"testing"

	"github.com/influxdata/influxdb/pkg/testing/assert"
)

func TestRegex(t *testing.T) {
	assert.Equal(t, GetNumber("sync15.ts"), 15)
	assert.Equal(t, GetNumber("sync1.ts"), 1)
	assert.Equal(t, GetNumber("sync10.ts"), 10)
}
