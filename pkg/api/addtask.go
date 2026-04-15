package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/lentyss/go-planner-server/pkg/db"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func checkDate(task *db.Task) error {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	}

	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return errors.New("дата представлена в неверном формате, нужный 20060102")
	}

	var nextDate string
	if task.Repeat != "" {
		nextDate, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	if now.After(t) {
		if task.Repeat == "" {
			task.Date = now.Format(DateFormat)
		} else {
			task.Date = nextDate
		}
	}
	return nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}
	defer r.Body.Close()

	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "не указан заголовок задачи"})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ошибка при сохранении задачи"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": formatInt64(id)})
}

func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
