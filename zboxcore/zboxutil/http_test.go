package zboxutil

import (
	coreHttp "github.com/0chain/gosdk/core/client"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCurrentDominantStatus(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		status                int
		runningTotalPerStatus map[int]int
		runningMax            int
		wantIsDominant        bool
	}{
		{
			name:                  "first response - 200",
			status:                200,
			runningTotalPerStatus: map[int]int{200: 1},
			runningMax:            1,
			wantIsDominant:        true,
		},
		{
			name:                  "first response - 500",
			status:                500,
			runningTotalPerStatus: map[int]int{500: 1},
			runningMax:            1,
			wantIsDominant:        true,
		},
		{
			name:                  "current response - 200 (previous was 500) - tiebreakers",
			status:                200,
			runningTotalPerStatus: map[int]int{200: 1, 500: 1},
			runningMax:            1,
			wantIsDominant:        true,
		},
		{
			name:                  "current response - 500 (previous was 200) - tiebreakers - should not be dominant",
			status:                500,
			runningTotalPerStatus: map[int]int{200: 1, 500: 1},
			runningMax:            1,
			wantIsDominant:        false,
		},
		{
			name:                  "current response - 500 (previous were 200, 500)",
			status:                500,
			runningTotalPerStatus: map[int]int{200: 1, 500: 2},
			runningMax:            2,
			wantIsDominant:        true,
		},
		{
			name:                  "current response - 200 (previous were 400, 404, 500) - tiebreakers",
			status:                200,
			runningTotalPerStatus: map[int]int{200: 1, 400: 1, 404: 1, 500: 1},
			runningMax:            1,
			wantIsDominant:        true,
		},
	} {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			got := coreHttp.IsCurrentDominantStatus(tt.status, tt.runningTotalPerStatus, tt.runningMax)

			assert.Equal(t, tt.wantIsDominant, got)
		})
	}
}
