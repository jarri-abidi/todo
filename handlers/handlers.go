package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jarri-abidi/todolist/todos"
)

type Handler struct {
	Service todos.Service
}

func (h *Handler) ToggleTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	err = h.Service.ToggleDone(id)
	if err == todos.ErrTodoNotFound {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) SaveTodo(w http.ResponseWriter, r *http.Request) {
	var t todos.Todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := h.Service.Save(&t); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, t)
}

func (h *Handler) RemoveTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	err = h.Service.Remove(id)
	if err == todos.ErrTodoNotFound {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ListTodos(w http.ResponseWriter, r *http.Request) {
	todolist, err := h.Service.List()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, todolist)
}

func (h *Handler) ReplaceTodo(w http.ResponseWriter, r *http.Request) {}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)

	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("could not encode http response: %v", err)
		return
	}

	if _, err := w.Write(response); err != nil {
		log.Printf("could not write http response: %v", err)
	}
}
