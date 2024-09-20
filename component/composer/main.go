package main

import (
	"net/http"

	"go.wasmcloud.dev/component/log/wasilog"
	"go.wasmcloud.dev/component/net/wasihttp"
)

const (
	componentName = "composer"
)

var logger = wasilog.ContextLogger("composer")

func init() {
	wasihttp.HandleFunc(handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling request", "request", r)
	switch r.Method {
	case http.MethodGet:
		readHandler(w, r)
	case http.MethodPost:
		createHandler(w, r)
	case http.MethodPut:
		updateHandler(w, r)
	case http.MethodDelete:
		deleteHandler(w, r)
	default:
		logger.Error("Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

//go:generate wit-bindgen-go generate --world function --out gen ./wit
func main() {}
