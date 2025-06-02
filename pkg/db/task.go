package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Task представляет структуру задачи
type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет новую задачу в базу данных
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) 
	          VALUES (?, ?, ?, ?)`

	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// Tasks возвращает список задач с поддержкой поиска
func Tasks(limit int, search string) ([]*Task, error) {
	var query string
	var args []interface{}

	switch {
	case isDateSearch(search):
		date, err := parseSearchDate(search)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		query = `SELECT id, date, title, comment, repeat 
		         FROM scheduler 
		         WHERE date = ? 
		         ORDER BY date ASC 
		         LIMIT ?`
		args = []interface{}{date, limit}

	case search != "":
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = `SELECT id, date, title, comment, repeat 
		         FROM scheduler 
		         WHERE LOWER(title) LIKE ? OR LOWER(comment) LIKE ? 
		         ORDER BY date ASC 
		         LIMIT ?`
		args = []interface{}{searchTerm, searchTerm, limit}

	default:
		query = `SELECT id, date, title, comment, repeat 
		         FROM scheduler 
		         ORDER BY date ASC 
		         LIMIT ?`
		args = []interface{}{limit}
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title,
			&task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Гарантируем возврат [] вместо nil
	if tasks == nil {
		tasks = make([]*Task, 0)
	}

	return tasks, nil
}

// isDateSearch проверяет, является ли строка датой в формате dd.mm.yyyy
func isDateSearch(s string) bool {
	_, err := time.Parse("02.01.2006", s)
	return err == nil
}

// parseSearchDate преобразует строку dd.mm.yyyy в формат yyyymmdd
func parseSearchDate(s string) (string, error) {
	t, err := time.Parse("02.01.2006", s)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}
	return t.Format("20060102"), nil
}

// DeleteTask удаляет задачу по ID
func DeleteTask(id int64) error {
	_, err := DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// GetTask возвращает задачу по ID
func GetTask(id int64) (*Task, error) {
	var task Task
	query := `SELECT id, date, title, comment, repeat 
	          FROM scheduler 
	          WHERE id = ?`

	err := DB.QueryRow(query, id).Scan(
		&task.ID,
		&task.Date,
		&task.Title,
		&task.Comment,
		&task.Repeat,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// UpdateTask обновляет существующую задачу
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler 
	          SET date = ?, title = ?, comment = ?, repeat = ? 
	          WHERE id = ?`

	res, err := DB.Exec(query,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
