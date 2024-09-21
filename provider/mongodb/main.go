package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	server "github.com/jamesstocktonj1/mulib/provider/mongodb/gen"
	"go.wasmcloud.dev/provider"
)

func run() error {
	// Create a new provider handler
	providerHandler := Handler{}

	// Create a new provider
	p, err := provider.New(
		provider.TargetLinkPut(providerHandler.handleNewTargetLink),
		provider.TargetLinkDel(providerHandler.handleDelTargetLink),
		provider.HealthCheck(providerHandler.handleHealthCheck),
		provider.Shutdown(providerHandler.handleShutdown),
	)
	if err != nil {
		return err
	}

	// Setup two channels to await RPC and control interface operations
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle RPC operations
	stopFunc, err := server.Serve(p.RPCClient, &providerHandler)
	if err != nil {
		p.Shutdown()
		return err
	}

	// Handle control interface operations
	go func() {
		err := p.Start()
		providerCh <- err
	}()

	// Shutdown on SIGINT
	signal.Notify(signalCh, syscall.SIGINT)

	// Run provider until either a shutdown is requested or a SIGINT is received
	select {
	case err = <-providerCh:
		stopFunc()
		return err
	case <-signalCh:
		p.Shutdown()
		stopFunc()
	}

	return nil
}

func (h *Handler) handleNewTargetLink(msg provider.InterfaceLinkDefinition) error {
	return nil
}

func (h *Handler) handleDelTargetLink(msg provider.InterfaceLinkDefinition) error {
	return nil
}

func (h *Handler) handleHealthCheck() string {
	return "Healthy"
}

func (h *Handler) handleShutdown() error {
	return nil
}

//go:generate wit-bindgen-wrpc go --out-dir gen --package github.com/jamesstocktonj1/mulib/provider/mongodb/gen wit
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
