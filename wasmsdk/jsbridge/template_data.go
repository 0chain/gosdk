//go:build js && wasm

package jsbridge

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall/js"
	"text/template"
)

//go:embed zcnworker.js.tpl
var WorkerJSTpl []byte

func buildWorkerJS(args, env []string, path string) (string, error) {
	return buildJS(args, env, path, WorkerJSTpl)
}

func buildJS(args, env []string, path string, tpl []byte) (string, error) {
	var workerJS bytes.Buffer

	if len(args) == 0 {
		args = []string{path}
	}

	if len(env) == 0 {
		env = os.Environ()
	}

	if uRL, err := url.ParseRequestURI(path); err != nil || !uRL.IsAbs() {
		origin := js.Global().Get("location").Get("origin").String()
		baseURL, err := url.ParseRequestURI(origin)
		if err != nil {
			return "", err
		}
		path = baseURL.JoinPath(path).String()
	}

	data := templateData{
		Path: path,
		Args: args,
		Env:  env,
	}
	if err := template.Must(template.New("js").Parse(string(tpl))).Execute(&workerJS, data); err != nil {
		return "", err
	}
	return workerJS.String(), nil
}

type templateData struct {
	Path string
	Args []string
	Env  []string
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
