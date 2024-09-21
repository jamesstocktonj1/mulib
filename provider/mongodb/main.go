package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	server "github.com/jamesstocktonj1/mulib/provider/mongodb/gen"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.wasmcloud.dev/provider"
)

func run() error {
	// Create a new provider handler
	providerHandler := Handler{
		clientMap: make(map[string]*HandlerConfig),
	}

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
	providerHandler.provider = p

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

func (h *Handler) handleNewTargetLink(link provider.InterfaceLinkDefinition) error {
	h.provider.Logger.Debug("Handling new target link", "link", link)
	var err error
	handlerConfig := HandlerConfig{
		config: link.SourceConfig,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client()
	clientOptions.ApplyURI(link.SourceConfig["uri"])
	h.provider.Logger.Debug("Connecting to MongoDB", "options", clientOptions)

	handlerConfig.client, err = mongo.Connect(ctx, options.Client().ApplyURI(link.SourceConfig["uri"]))
	if err != nil {
		h.provider.Logger.Error("Error connecting to MongoDB", "error", err)
		return err
	}

	h.clientMap[link.SourceID] = &handlerConfig
	return nil
}

func (h *Handler) handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	h.provider.Logger.Debug("Handling delete target link", "link", link)

	delete(h.clientMap, link.SourceID)
	return nil
}

func (h *Handler) handleHealthCheck() string {
	h.provider.Logger.Debug("Handling health check")
	health := provider.HealthCheckResponse{
		Healthy: true,
	}

	for _, c := range h.clientMap {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.client.Ping(ctx, nil); err != nil {
			health.Healthy = false
			health.Message = err.Error()
			break
		}
	}

	data, err := json.Marshal(health)
	if err != nil {
		return "unable to marshal health check response"
	}
	return string(data)
}

func (h *Handler) handleShutdown() error {
	h.provider.Logger.Debug("Handling shutdown")
	return nil
}

//go:generate wit-bindgen-wrpc go --out-dir gen --package github.com/jamesstocktonj1/mulib/provider/mongodb/gen wit
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
