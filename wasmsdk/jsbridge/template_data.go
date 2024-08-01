//go:build js && wasm

package jsbridge

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"syscall/js"
	"text/template"

	"github.com/0chain/gosdk/core/version"
)

//go:embed zcnworker.js.tpl
var WorkerJSTpl []byte

func buildWorkerJS(args, env []string, path string) (string, error) {
	return buildJS(args, env, path, WorkerJSTpl)
}

func buildJS(args, env []string, wasmPath string, tpl []byte) (string, error) {
	var workerJS bytes.Buffer

	if len(args) == 0 {
		args = []string{wasmPath}
	}

	if len(env) == 0 {
		env = os.Environ()
	}
	var cachePath string
	if uRL, err := url.ParseRequestURI(wasmPath); err != nil || !uRL.IsAbs() {
		origin := js.Global().Get("location").Get("origin").String()
		u, err := url.Parse(origin)
		if err != nil {
			return "", err
		}
		u.Path = path.Join(u.Path, wasmPath)
		cachePath = u.String()
		params := url.Values{}
		params.Add("v", version.VERSIONSTR)
		u.RawQuery = params.Encode()
		wasmPath = u.String()
	}
	cdnPath := "https://d2os1u2xwjukgr.cloudfront.net/dev/zcn.wasm"
	data := templateData{
		Path:         cdnPath,
		Args:         args,
		Env:          env,
		FallbackPath: wasmPath,
		CachePath:    cachePath,
	}
	if err := template.Must(template.New("js").Parse(string(tpl))).Execute(&workerJS, data); err != nil {
		return "", err
	}
	return workerJS.String(), nil
}

type templateData struct {
	Path         string
	Args         []string
	Env          []string
	FallbackPath string
	CachePath    string
}

func (d templateData) ArgsToJS() string {
	el := []string{}
	for _, e := range d.Args {
		el = append(el, `"`+e+`"`)
	}
	return "[" + strings.Join(el, ",") + "]"
}

func (d templateData) EnvToJS() string {
	el := []string{}
	for _, entry := range d.Env {
		if k, v, ok := strings.Cut(entry, "="); ok {
			el = append(el, fmt.Sprintf(`"%s":"%s"`, k, v))
		}
	}
	return "{" + strings.Join(el, ",") + "}"
}
