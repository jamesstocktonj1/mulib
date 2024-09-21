package main

import (
	"context"

	"github.com/jamesstocktonj1/mulib/provider/mongodb/gen/exports/wasmcloud/document/document"
	wrpc "wrpc.io/go"
)

// Handler is the MongoDB provider implementation
type Handler struct{}

// ensure Handler implements the document.Handler interface
var _ document.Handler = &Handler{}

func (h *Handler) Insert(ctx__ context.Context, col []uint8, doc *document.Document) (*wrpc.Result[struct{}, string], error) {
	return wrpc.Ok[string](struct{}{}), nil
}

func (h *Handler) Find(ctx__ context.Context, col []uint8, filter *document.Filter) (*wrpc.Result[[]*document.Document, string], error) {
	return wrpc.Ok[string]([]*document.Document{}), nil
}

func (h *Handler) Update(ctx__ context.Context, col []uint8, doc *document.Document, filter *document.Filter) (*wrpc.Result[struct{}, string], error) {
	return wrpc.Ok[string](struct{}{}), nil
}

func (h *Handler) Delete(ctx__ context.Context, col []uint8, filter *document.Filter) (*wrpc.Result[struct{}, string], error) {
	return wrpc.Ok[string](struct{}{}), nil
}
