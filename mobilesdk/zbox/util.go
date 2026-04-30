package zbox

import (
	"regexp"
	"strconv"

	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// SetSessionID sets the session/device ID on all outgoing blobber HTTP
// requests. Blobbers forward this to 0box for activity-sync event
// attribution and own-session suppression.
// Mirrors wasmsdk/wallet.go:setSessionID — calls both resty (for requests
// going through the resty transport) and zboxutil (for direct http requests).
func SetSessionID(id string) {
	resty.SetSessionID(id)
	zboxutil.SetSessionID(id)
}

func GetNumber(value string) int {
	re := regexp.MustCompile("[0-9]+")
	submatchall := re.FindAllString(value, -1)
	for _, element := range submatchall {
		res, _ := strconv.Atoi(element)
		return res
	}
	return -1
}
