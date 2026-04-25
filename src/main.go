package main

import "fmt"

func main() {
    db := InitDB()
    defer db.Close()

    SeedMasterData(db)
}
