package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"kakeibo/internal/db"
	"kakeibo/internal/service/api"
	"kakeibo/internal/service/api/dto"
	"kakeibo/internal/service/csv"
	"kakeibo/internal/service/setup"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

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
	if err := importCSV(conn); err != nil {
		log.Fatal(err)
	}
	srv := api.NewServer(conn)
	if os.Getenv("SEED_TEST_TRANSACTIONS") == "1" || os.Getenv("SEED_TEST_TRANSACTIONS") == "true" {
		if err := seedTestTransactionsViaAPI(srv); err != nil {
			log.Fatal(err)
		}
	}
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

func seedTestTransactionsViaAPI(srv *api.Server) error {
	h := srv.Routes()

	requests := []dto.CreateTransactionRequest{
		{
			UserID:     1,
			Date:       "2026-06-20",
			Amount:     2000,
			NetAmount:  500,
			Type:       false,
			IsTransfer: false,
			Place:      "テストスーパー",
			Note:       "test data: lunch",
			MethodID:   1,
			CategoryID: nil,
			Advances: []dto.AdvanceInput{
				{
					Name:   "溝口洸熙",
					Amount: 1000,
				},
				{
					Name:   "溝口洸熙2",
					Amount: 500,
				},
			},
			RefundAdvanceID: nil,
		},
		{
			UserID:          1,
			Date:            "2026-06-21",
			Amount:          3000,
			NetAmount:       3000,
			Type:            true,
			IsTransfer:      false,
			Place:           "テスト口座",
			Note:            "test data: salary",
			MethodID:        1,
			CategoryID:      nil,
			Advances:        []dto.AdvanceInput{},
			RefundAdvanceID: nil,
		},
	}

	for i, reqBody := range requests {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			body, _ := io.ReadAll(rr.Result().Body)
			return fmt.Errorf("seed transaction %d failed: status=%d body=%s", i+1, rr.Code, string(body))
		}
	}

	log.Printf("✅ seeded %d test transactions via API", len(requests))

	return nil
}
