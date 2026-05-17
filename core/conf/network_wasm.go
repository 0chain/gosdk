//go:build js && wasm

package conf

import (
	"net/url"
)

// rewriteURLsForBrowser converts HTTP miner/sharder URLs to HTTPS reverse
// proxy URLs to avoid mixed-content blocking in browsers. All mainnet
// miners expose /miner01 and sharders expose /sharder01 HTTPS reverse proxies.
func rewriteURLsForBrowser(urls []string) []string {
	for i, raw := range urls {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme == "https" {
			continue
		}
		port := parsed.Port()
		switch port {
		case "7171":
			parsed.Scheme = "https"
			parsed.Host = parsed.Hostname()
			parsed.Path = "/sharder01"
		case "7071":
			parsed.Scheme = "https"
			parsed.Host = parsed.Hostname()
			parsed.Path = "/miner01"
		default:
			continue
		}
		urls[i] = parsed.String()
	}
	return urls
}
