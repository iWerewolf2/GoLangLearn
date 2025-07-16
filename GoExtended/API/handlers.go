package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Index responds with a welcome message.
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

// TodoIndex responds with all todos in JSON.
func TodoIndex(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, todos)
}

// TodoShow responds with a specific todo by ID or 404 if not found.
func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoID, err := strconv.Atoi(vars["todoId"])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, jsonErr{Code: http.StatusBadRequest, Text: "Invalid ID"})
		return
	}

	todo := RepoFindTodo(todoID)
	if todo.Id > 0 {
		writeJSON(w, http.StatusOK, todo)
	} else {
		writeJSON(w, http.StatusNotFound, jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
	}
}

// TodoCreate creates a new todo from the request body.
func TodoCreate(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, jsonErr{Code: http.StatusInternalServerError, Text: "Failed to read request"})
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &todo); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, jsonErr{Code: http.StatusUnprocessableEntity, Text: "Invalid JSON"})
		return
	}

	createdTodo := RepoCreateTodo(todo)
	writeJSON(w, http.StatusCreated, createdTodo)
}

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
