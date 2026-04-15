package api

import (
	"encoding/json"
	"net/http"

	"github.com/lentyss/go-planner-server/pkg/db"
)

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ошибка десериализации и JSON"})
		return
	}
	defer r.Body.Close()

	if task.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "не указан идентификатор"})
		return
	}

	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "не указан заголовок"})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}
