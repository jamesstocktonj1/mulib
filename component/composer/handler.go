package main

import (
	"encoding/json"
	"net/http"

	"github.com/bytecodealliance/wasm-tools-go/cm"
	"github.com/google/uuid"
	"github.com/jamesstocktonj1/mulib/component/composer/gen/wasi/keyvalue/store"
	"github.com/jamesstocktonj1/mulib/pkg/composer"
)

const (
	ErrBadRequest = "err:"
)

func createHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Creating new composer")

	// Unmarshal request
	comp := composer.Composer{}
	err := json.NewDecoder(r.Body).Decode(&comp)
	if err != nil {
		logger.Error("Error decoding request", "error", err)
		http.Error(w, "error decoding request", http.StatusBadRequest)
		return
	}

	// Open bucket
	bucketRes := store.Open(componentName)
	if bucketRes.IsErr() {
		logger.Error("Error opening bucket", "error", bucketRes.Err())
		http.Error(w, "error opening bucket", http.StatusInternalServerError)
		return
	}
	bucket := bucketRes.OK()

	// Set ID
	comp.ID = uuid.New().String()

	// Check if value exists
	existsRes := bucket.Exists(comp.ID)
	if existsRes.IsErr() {
		logger.Error("Error checking if value exists", "error", existsRes.Err())
		http.Error(w, "error checking if value exists", http.StatusInternalServerError)
		return
	} else if *existsRes.OK() {
		logger.Error("Value already exist", "id", comp.ID, "result", *existsRes.OK())
		http.Error(w, "value already exist", http.StatusNotFound)
		return
	}

	// Marshal value
	compBytes, err := json.Marshal(comp)
	if err != nil {
		logger.Error("Error marshalling value", "error", err)
		http.Error(w, "error marshalling value", http.StatusInternalServerError)
		return
	}

	// Set value
	res := bucket.Set(comp.ID, cm.ToList(compBytes))
	if res.IsErr() {
		logger.Error("Error setting value", "error", res.Err())
		http.Error(w, "error setting value", http.StatusInternalServerError)
		return
	}

	// Write response
	idResponse := map[string]string{
		"id":      comp.ID,
		"message": "composer created",
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(idResponse)
	if err != nil {
		logger.Error("Error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Reading composer")

	// Get composer ID
	id := r.URL.Query().Get("composer")
	if id == "" {
		logger.Error("No composer query provided")
		http.Error(w, "no composer query provided", http.StatusBadRequest)
		return
	}

	// Open bucket
	bucketRes := store.Open(componentName)
	if bucketRes.IsErr() {
		logger.Error("Error opening bucket", "error", bucketRes.Err())
		http.Error(w, "error opening bucket", http.StatusInternalServerError)
		return
	}
	bucket := bucketRes.OK()

	// Check if value exists
	existsRes := bucket.Exists(id)
	if existsRes.IsErr() {
		logger.Error("Error checking if value exists", "error", existsRes.Err())
		http.Error(w, "error checking if value exists", http.StatusInternalServerError)
		return
	} else if !(*existsRes.OK()) {
		logger.Error("Value does not exist", "id", id, "result", *existsRes.OK())
		http.Error(w, "value does not exist", http.StatusNotFound)
		return
	}

	// Get value
	res := bucket.Get(id)
	if res.IsErr() {
		logger.Error("Error getting value", "error", res.Err())
		http.Error(w, "error reading value", http.StatusInternalServerError)
		return
	}

	// Unmarshal value
	comp := composer.Composer{}
	err := json.Unmarshal(res.OK().Some().Slice(), &comp)
	if err != nil {
		logger.Error("Error unmarshalling bucket.Get value", "error", err)
		http.Error(w, "error reading value", http.StatusInternalServerError)
		return
	}

	// Marshal response
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(comp)
	if err != nil {
		logger.Error("Error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Updating composer")

	// Get composer ID
	id := r.URL.Query().Get("composer")
	if id == "" {
		logger.Error("No composer query provided")
		http.Error(w, "no composer query provided", http.StatusBadRequest)
		return
	}

	// Unmarshal request
	compPut := composer.Composer{}
	err := json.NewDecoder(r.Body).Decode(&compPut)
	if err != nil {
		logger.Error("Error decoding request", "error", err)
		http.Error(w, "error decoding request", http.StatusBadRequest)
		return
	}

	// Open bucket
	bucketRes := store.Open(componentName)
	if bucketRes.IsErr() {
		logger.Error("Error opening bucket", "error", bucketRes.Err())
		http.Error(w, "error opening bucket", http.StatusInternalServerError)
		return
	}
	bucket := bucketRes.OK()

	// Check if value exists
	existsRes := bucket.Exists(id)
	if existsRes.IsErr() {
		logger.Error("Error checking if value exists", "error", existsRes.Err())
		http.Error(w, "error checking if value exists", http.StatusInternalServerError)
		return
	} else if !(*existsRes.OK()) {
		logger.Error("Value does not exist", "id", id, "result", *existsRes.OK())
		http.Error(w, "value does not exist", http.StatusNotFound)
		return
	}

	// Get value
	res := bucket.Get(id)
	if res.IsErr() {
		logger.Error("Error getting value", "error", res.Err())
		http.Error(w, "error reading value", http.StatusInternalServerError)
		return
	}

	// Unmarshal value
	comp := composer.Composer{}
	err = json.Unmarshal(res.OK().Some().Slice(), &comp)
	if err != nil {
		logger.Error("Error unmarshalling bucket.Get value", "error", err)
		http.Error(w, "error reading value", http.StatusInternalServerError)
		return
	}

	// Update value
	if compPut.Firstname != "" {
		comp.Firstname = compPut.Firstname
	}
	if compPut.Lastname != "" {
		comp.Lastname = compPut.Lastname
	}
	if compPut.BirthDate != "" {
		comp.BirthDate = compPut.BirthDate
	}
	if compPut.DeathDate != "" {
		comp.DeathDate = compPut.DeathDate
	}
	if compPut.Era != "" {
		comp.Era = compPut.Era
	}
	if compPut.Nationality != "" {
		comp.Nationality = compPut.Nationality
	}

	// Marshal value
	compBytes, err := json.Marshal(comp)
	if err != nil {
		logger.Error("Error marshalling value", "error", err)
		http.Error(w, "error marshalling value", http.StatusInternalServerError)
		return
	}

	// Set value
	setRes := bucket.Set(id, cm.ToList(compBytes))
	if setRes.IsErr() {
		logger.Error("Error setting value", "error", setRes.Err())
		http.Error(w, "error setting value", http.StatusInternalServerError)
		return
	}

	// Marshal response
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(comp)
	if err != nil {
		logger.Error("Error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Deleting composer")

	// Get composer ID
	id := r.URL.Query().Get("composer")
	if id == "" {
		logger.Error("No composer query provided")
		http.Error(w, "no composer query provided", http.StatusBadRequest)
		return
	}

	// Open bucket
	bucketRes := store.Open(componentName)
	if bucketRes.IsErr() {
		logger.Error("Error opening bucket", "error", bucketRes.Err())
		http.Error(w, "error opening bucket", http.StatusInternalServerError)
		return
	}
	bucket := bucketRes.OK()

	// Check if value exists
	existsRes := bucket.Exists(id)
	if existsRes.IsErr() {
		logger.Error("Error checking if value exists", "error", existsRes.Err())
		http.Error(w, "error checking if value exists", http.StatusInternalServerError)
		return
	} else if !(*existsRes.OK()) {
		logger.Error("Value does not exist", "id", id, "result", *existsRes.OK())
		http.Error(w, "value does not exist", http.StatusNotFound)
		return
	}

	// Delete value
	res := bucket.Delete(id)
	if res.IsErr() {
		logger.Error("Error deleting value", "error", res.Err())
		http.Error(w, "error deleting value", http.StatusInternalServerError)
		return
	}

	// Write response
	idResponse := map[string]string{
		"id":      id,
		"message": "composer deleted",
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(idResponse)
	if err != nil {
		logger.Error("Error encoding response", "error", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
