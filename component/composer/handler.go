package main

import (
	"net/http"
)

const (
	ErrBadRequest = "err:"
)

func createHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Creating new composer")

	w.Write([]byte("Created composer"))
	w.WriteHeader(http.StatusCreated)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Reading composer")

	w.Write([]byte("Reading composer"))
	w.WriteHeader(http.StatusOK)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Updating composer")

	w.Write([]byte("Updated composer"))
	w.WriteHeader(http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Deleting composer")

	w.Write([]byte("Deleted composer"))
	w.WriteHeader(http.StatusOK)
}
