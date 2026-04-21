package main

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./highwaycruizzers.db")
    if err != nil {
        log.Fatal("Error:", err)
    }
    defer db.Close()
    
    var count int
    db.QueryRow("SELECT COUNT(*) FROM points_of_interest").Scan(&count)
    log.Printf("POI count: %d", count)
    
    // Show first few records
    rows, err := db.Query("SELECT id, name, type FROM points_of_interest LIMIT 5")
    if err == nil {
        defer rows.Close()
        for rows.Next() {
            var id int
            var name, poiType string
            rows.Scan(&id, &name, &poiType)
            log.Printf("ID: %d, Name: %s, Type: %s", id, name, poiType)
        }
    }
}