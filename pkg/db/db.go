package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

const (
	defaultDBFile = "scheduler.db"
	schema        = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`
)

var (
	dbConn *sqlx.DB // Новое соединение
)

// SetDBConnection устанавливает соединение (и для sql.DB и для sqlx.DB)
func SetDBConnection(conn *sqlx.DB) {
	dbConn = conn
	DB = conn.DB // Получаем стандартное соединение из sqlx
}

// Init инициализирует базу данных
func Init() error {
	dbFile := getDBFilePath()

	_, err := os.Stat(dbFile)
	install := errors.Is(err, os.ErrNotExist)

	if dir := filepath.Dir(dbFile); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if install {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	DB = db
	return nil
}

func getDBFilePath() string {
	if dbFile := os.Getenv("TODO_DBFILE"); dbFile != "" {
		return dbFile
	}
	return defaultDBFile
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) 
	          VALUES (?, ?, ?, ?)`

	log.Printf("Adding task: Date=%s, Title=%s, Comment=%s, Repeat=%s",
		task.Date, task.Title, task.Comment, task.Repeat)

	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	log.Printf("Task added successfully, ID=%d", id)
	return id, nil
}

func UpdateTaskDate(id int64, date string) error {
	_, err := DB.Exec(`
        UPDATE scheduler 
        SET date = ?
        WHERE id = ?`,
		date, id,
	)
	return err
}
