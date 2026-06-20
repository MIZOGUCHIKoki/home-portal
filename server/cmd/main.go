package main

import (
	"database/sql"
	"kakeibo/internal/db"
	"kakeibo/internal/service/api"
	"kakeibo/internal/service/csv"
	"kakeibo/internal/service/setup"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	conn, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := setup.Run(conn); err != nil {
		log.Fatal(err)
	}
	// if err := importCSV(conn); err != nil {
	// 	log.Fatal(err)
	// }
	srv := api.NewServer(conn)
	srv.Run(":8080")

}

func importCSV(conn *sql.DB) error {

	if err := csv.ImportJCB(conn, "assets/csv/jcb", 1); err != nil {
		return err
	}
	if err := csv.ImportSBINetBank(conn, "assets/csv/netbk", 1); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
