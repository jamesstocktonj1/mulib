package main

import (
	"context"
	"errors"

	"github.com/jamesstocktonj1/mulib/provider/mongodb/gen/exports/wasmcloud/document/document"
	"github.com/jamesstocktonj1/mulib/provider/mongodb/gen/wasmcloud/document/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.wasmcloud.dev/provider"
	wrpc "wrpc.io/go"
	wrpcnats "wrpc.io/go/nats"
)

var (
	ErrNoHeadersInContext = errors.New("no headers in context")
	ErrSourceNotLinked    = errors.New("source not linked")
	ErrClientNil          = errors.New("source linked but client is nil")
)

// Handler is the MongoDB provider implementation
type Handler struct {
	provider  *provider.WasmcloudProvider
	clientMap map[string]*HandlerConfig
}

type HandlerConfig struct {
	config map[string]string
	client *mongo.Client
}

// ensure Handler implements the document.Handler interface
var _ document.Handler = &Handler{}

func (h *Handler) Insert(ctx__ context.Context, col []uint8, doc *document.Document) (*wrpc.Result[struct{}, string], error) {
	handler, err := h.getHandlerConfig(ctx__)
	if err != nil {
		h.provider.Logger.Error("failed to getHandlerConfig", "error", err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}

	bsonDoc := bson.M{}
	err = bson.UnmarshalExtJSON(doc.Data, true, &bsonDoc)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}
	h.provider.Logger.Debug("bson document", "doc", bsonDoc)

	collection := handler.getDatabase().Collection(string(col))
	res, err := collection.InsertOne(ctx__, bsonDoc)
	if err != nil {
		h.provider.Logger.Error("failed to insert to collection", "error", err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}
	h.provider.Logger.Debug("inserted document", "id", res.InsertedID)

	return wrpc.Ok[string](struct{}{}), nil
}

func (h *Handler) Find(ctx__ context.Context, col []uint8, filter *document.Filter) (*wrpc.Result[[]*document.Document, string], error) {
	handler, err := h.getHandlerConfig(ctx__)
	if err != nil {
		h.provider.Logger.Error("failed to getHandlerConfig", "error", err.Error())
		return wrpc.Err[[]*types.Document](err.Error()), err
	}

	bsonFilter := bson.D{}
	err = bson.UnmarshalExtJSON(filter.Data, true, &bsonFilter)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[[]*types.Document](err.Error()), err
	}
	h.provider.Logger.Debug("bson filter", "filter", bsonFilter)

	collection := handler.getDatabase().Collection(string(col))
	cursor, err := collection.Find(ctx__, bsonFilter)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[[]*types.Document](err.Error()), err
	}

	var results []bson.M
	if err = cursor.All(ctx__, &results); err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[[]*types.Document](err.Error()), err
	}

	docs := make([]*types.Document, len(results))
	for i, result := range results {
		data, err := bson.MarshalExtJSON(result, false, true)
		if err != nil {
			h.provider.Logger.Error(err.Error())
		} else {
			docs[i] = &types.Document{Data: data}
		}
	}
	return wrpc.Ok[string](docs), nil
}

func (h *Handler) Update(ctx__ context.Context, col []uint8, doc *document.Document, filter *document.Filter) (*wrpc.Result[struct{}, string], error) {
	handler, err := h.getHandlerConfig(ctx__)
	if err != nil {
		h.provider.Logger.Error("failed to getHandlerConfig", "error", err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}

	bsonFilter := bson.D{}
	err = bson.UnmarshalExtJSON(filter.Data, true, &bsonFilter)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}
	h.provider.Logger.Debug("bson filter", "filter", bsonFilter)

	bsonDoc := bson.M{}
	err = bson.UnmarshalExtJSON(doc.Data, true, &bsonDoc)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}
	h.provider.Logger.Debug("bson document", "doc", bsonDoc)

	collection := handler.getDatabase().Collection(string(col))
	res, err := collection.UpdateMany(ctx__, bsonFilter, bsonDoc)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}
	h.provider.Logger.Debug("updated documents", "count", res.ModifiedCount, "ids", res.UpsertedID)

	if res.MatchedCount == 0 {
		h.provider.Logger.Warn("no documents matched the filter")
	}
	return wrpc.Ok[string](struct{}{}), nil
}

func (h *Handler) Delete(ctx__ context.Context, col []uint8, filter *document.Filter) (*wrpc.Result[struct{}, string], error) {
	handler, err := h.getHandlerConfig(ctx__)
	if err != nil {
		h.provider.Logger.Error("failed to getHandlerConfig", "error", err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}

	bsonFilter := bson.D{}
	err = bson.UnmarshalExtJSON(filter.Data, true, &bsonFilter)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}
	h.provider.Logger.Debug("bson filter", "filter", bsonFilter)

	collection := handler.getDatabase().Collection(string(col))
	res, err := collection.DeleteMany(ctx__, bsonFilter)
	if err != nil {
		h.provider.Logger.Error(err.Error())
		return wrpc.Err[struct{}](err.Error()), err
	}

	if res.DeletedCount == 0 {
		h.provider.Logger.Warn("no documents matched the filter")
	}
	return wrpc.Ok[string](struct{}{}), nil
}

func (h *Handler) getHandlerConfig(ctx context.Context) (*HandlerConfig, error) {
	headers, ok := wrpcnats.HeaderFromContext(ctx)
	if !ok {
		return nil, ErrNoHeadersInContext
	}

	sourceID := headers.Get("source-id")
	client, ok := h.clientMap[sourceID]
	if !ok || client == nil {
		return nil, ErrSourceNotLinked
	}
	return client, nil
}

func (c *HandlerConfig) getDatabase() *mongo.Database {
	databaseName := "default"
	if name, ok := c.config["database"]; ok {
		databaseName = name
	}
	return c.client.Database(databaseName)
}
