package db

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// ...
func InitDB() {
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	table := os.Getenv("DB_TABLE")

	dataSource := username + ":" + password + "@tcp(localhost:3306)/" + table

	var err error
	DB, err = sql.Open("mysql", dataSource)

	if err != nil {
		panic(err)
	}

	DB.SetConnMaxLifetime(time.Minute * 3)
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(10)
}
