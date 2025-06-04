package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

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
	if task.ID == 0 {
		return fmt.Errorf("task ID is required")
	}
	if task.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
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

func NextDate(currentDate, repeatRule string) (string, error) {
	if repeatRule == "" {
		return "", fmt.Errorf("empty repeat rule")
	}

	currentTime, err := time.Parse("20060102", currentDate)
	if err != nil {
		return "", fmt.Errorf("invalid current date format: %v", err)
	}

	parts := strings.Fields(repeatRule)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid repeat rule format")
	}

	switch parts[0] {
	case "d": // Ежедневное повторение
		return dailyRepeat(currentTime, parts[1])
	case "w": // Еженедельное повторение
		return weeklyRepeat(currentTime, parts[1])
	case "m": // Ежемесячное повторение
		return monthlyRepeat(currentTime, parts[1])
	case "y": // Ежегодное повторение
		return yearlyRepeat(currentTime, parts[1])
	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", parts[0])
	}
}

// dailyRepeat вычисляет следующую дату для ежедневного повторения
func dailyRepeat(currentTime time.Time, interval string) (string, error) {
	days, err := strconv.Atoi(interval)
	if err != nil {
		return "", fmt.Errorf("invalid days interval: %v", err)
	}
	if days <= 0 {
		return "", fmt.Errorf("days interval must be positive")
	}
	return currentTime.AddDate(0, 0, days).Format("20060102"), nil
}

// weeklyRepeat вычисляет следующую дату для еженедельного повторения
func weeklyRepeat(currentTime time.Time, interval string) (string, error) {
	weeks, err := strconv.Atoi(interval)
	if err != nil {
		return "", fmt.Errorf("invalid weeks interval: %v", err)
	}
	if weeks <= 0 {
		return "", fmt.Errorf("weeks interval must be positive")
	}
	return currentTime.AddDate(0, 0, 7*weeks).Format("20060102"), nil
}

// monthlyRepeat вычисляет следующую дату для ежемесячного повторения
func monthlyRepeat(currentTime time.Time, interval string) (string, error) {
	months, err := strconv.Atoi(interval)
	if err != nil {
		return "", fmt.Errorf("invalid months interval: %v", err)
	}
	if months <= 0 {
		return "", fmt.Errorf("months interval must be positive")
	}

	// Сохраняем день месяца (если он есть в следующем месяце)
	nextDate := currentTime.AddDate(0, months, 0)

	// Если день месяца превышает количество дней в следующем месяце,
	// берем последний день месяца
	if currentTime.Day() != nextDate.Day() {
		nextDate = time.Date(nextDate.Year(), nextDate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	}

	return nextDate.Format("20060102"), nil
}

// yearlyRepeat вычисляет следующую дату для ежегодного повторения
func yearlyRepeat(currentTime time.Time, interval string) (string, error) {
	years, err := strconv.Atoi(interval)
	if err != nil {
		return "", fmt.Errorf("invalid years interval: %v", err)
	}
	if years <= 0 {
		return "", fmt.Errorf("years interval must be positive")
	}
	return currentTime.AddDate(years, 0, 0).Format("20060102"), nil
}

func MarkTaskDone(id int64) error {
	// Получаем задачу из БД
	task, err := GetTask(id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Для непериодических задач просто удаляем
	if task.Repeat == "" {
		if err := DeleteTask(id); err != nil {
			return fmt.Errorf("failed to delete task: %w", err)
		}
		return nil
	}

	// Для периодических задач вычисляем следующую дату
	nextDate, err := NextDate(task.Date, task.Repeat)
	if err != nil {
		return fmt.Errorf("failed to calculate next date: %w", err)
	}

	// Обновляем дату задачи
	if err := UpdateTaskDate(id, nextDate); err != nil {
		return fmt.Errorf("failed to update task date: %w", err)
	}

	return nil
}
