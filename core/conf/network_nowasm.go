//go:build !(js && wasm)

package conf

func rewriteURLsForBrowser(urls []string) []string {
	return urls
}
