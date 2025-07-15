package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	dsn := "root@tcp(127.0.0.1:3306)/webcrawler"

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening DB connection: %s", err)
	}

	// Ping the database to verify connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Error pinging DB: %s", err)
	}

	fmt.Println("✅ Connected to MySQL successfully!")
}
