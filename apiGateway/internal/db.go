package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// InsertDb осуществляет подключение к бд и загрузку туда первичных данных
func InsertDb(act string) int {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	connStr := fmt.Sprintf("postgres://%s:%s@postgres:5432/aiAsyncSystem?sslmode=disable", dbUser, dbPassword)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("Ошибка подключения к БД:", err)
		return 0
	}
	defer db.Close()

	var id int

	err = db.QueryRow("INSERT INTO tasks (action, status) VALUES ($1, $2) RETURNING id", act, "pending").Scan(&id)
	if err != nil {
		log.Println("Ошибка вставки в БД:", err)
		return 0
	}

	return id
}

// GetTaskStatus загружает статус задачи по id
func GetTaskStatus(id string) (string, error) {
	var status string
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	connStr := fmt.Sprintf("postgres://%s:%s@postgres:5432/aiAsyncSystem?sslmode=disable", dbUser, dbPassword)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.QueryRow("SELECT status FROM tasks WHERE id = $1", id).Scan(&status)
	if err != nil {
		return "", err
	}

	return status, nil
}
