package internal

import (
	"encoding/json"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

// HandleTask обслуживает адрес /task
func HandleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id := InsertDb(req.Action)
	if id == 0 {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	task := map[string]interface{}{
		"id":       id,
		"priority": req.Priority,
		"action":   req.Action,
		"payload":  req.Payload,
	}

	body, _ := json.Marshal(task)

	err = Channel.Publish(
		"", Queue.Name, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		http.Error(w, "Failed to publish task", http.StatusInternalServerError)
		return
	}

	response := map[string]int{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.Printf(" [x] Sent request: %s", req.Action)
}

// HandleStatus обслуживает адрес /status
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	status, err := GetTaskStatus(id)
	if err != nil {
		http.Error(w, "Failed to get task status", http.StatusNotFound)
		return
	}

	answer := map[string]string{
		"id":     id,
		"status": status,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		http.Error(w, "Failed to encode answer", http.StatusInternalServerError)
		return
	}
}
