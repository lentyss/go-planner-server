package api

import (
	"net/http"

	"github.com/lentyss/go-planner-server/pkg/db"
)

type TasksResponse struct {
	Tasks []db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")

	limit := 50

	tasks, err := db.Tasks(limit, search)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ошибка получения задач"})
		return
	}

	writeJSON(w, http.StatusOK, TasksResponse{Tasks: tasks})
}
