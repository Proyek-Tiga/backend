package config


import (
  "database/sql"
  "log"
  "os"


  // "github.com/joho/godotenv"
  _ "github.com/lib/pq"
)


var DB *sql.DB


func InitDB() {
  var err error


  // Load environment variables from .env file
  // err = godotenv.Load()
  // if err != nil {
    // log.Fatal("Error loading .env file")
  // }


  // Get environment variables
  dburl := os.Getenv("DB_URL")


  // Build connection string
  connStr := dburl


  // Open connection to database
  DB, err = sql.Open("postgres", connStr)
  if err != nil {
    log.Fatal(err)
  }


  // Ping database to ensure connection is established
  err = DB.Ping()
  if err != nil {
    log.Fatal(err)
  }


  log.Println("Database connected")
}
