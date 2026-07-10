package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"kakeibo/internal/model"
)

func ImportTransactions(
	reader io.Reader,
	userID int64,
	methodID int64,
) ([]model.Transaction, error) {

	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true

	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) <= 1 {
		return nil, errors.New("csv has no data rows")
	}

	var transactions []model.Transaction

	for i, row := range rows[1:] {
		line := i + 2 // ヘッダー分ずらす

		if len(row) < 5 {
			return nil, fmt.Errorf("line %d: invalid column count", line)
		}

		t, err := parseRow(row, userID, methodID)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

func parseRow(
	row []string,
	userID int64,
	methodID int64,
) (model.Transaction, error) {

	date, err := time.Parse("2006-01-02", row[0])
	if err != nil {
		return model.Transaction{}, errors.New("invalid date format")
	}

	place := strings.TrimSpace(row[1])

	income, err := strconv.Atoi(row[2])
	if err != nil {
		return model.Transaction{}, errors.New("invalid income")
	}

	expense, err := strconv.Atoi(row[3])
	if err != nil {
		return model.Transaction{}, errors.New("invalid expense")
	}

	if income > 0 && expense > 0 {
		return model.Transaction{}, errors.New("income and expense cannot both be > 0")
	}

	isTransfer := parseBool(row[4])

	var amount int
	var ttype bool // true = income, false = expense

	if income > 0 {
		amount = income
		ttype = true
	} else {
		amount = expense
		ttype = false
	}

	if amount <= 0 {
		return model.Transaction{}, errors.New("amount must be positive")
	}

	var placePtr *string
	if place != "" {
		v := place
		placePtr = &v
	}

	return model.Transaction{
		UserID:     userID,
		Date:       date,
		Amount:     amount,
		NetAmount:  amount,
		Type:       ttype,
		IsTransfer: isTransfer,
		MethodID:   methodID,
		Place:      placePtr,
	}, nil
}

func parseBool(s string) bool {
	switch strings.TrimSpace(s) {
	case "はい", "true", "1", "yes":
		return true
	default:
		return false
	}
}
