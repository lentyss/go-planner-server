package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetTask(id string) (*Task, error) {
	var task Task
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	err := DB.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, err
	}
	return &task, nil
}

func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func UpdateDate(id string, date string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, date, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("заача не найдена")
	}

	return nil
}

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задачи не найденго")
	}
	return nil
}

func Tasks(limit int, search string) ([]Task, error) {
	var rows *sql.Rows
	var err error

	if search == "" {
		query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`
		rows, err = DB.Query(query, limit)
	} else {
		if t, parseErr := time.Parse("02.01.2006", search); parseErr == nil {
			dateStr := t.Format("20060102")
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`
			rows, err = DB.Query(query, dateStr, limit)
		} else {
			searchPattern := "%" + search + "%"
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`
			rows, err = DB.Query(query, searchPattern, searchPattern, limit)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, nil
}

func DateOnlyAfter(date, now time.Time) bool {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	nowOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return dateOnly.After(nowOnly)
}
