package repository

import (
	"database/sql"
	"log"

	"kakeibo/internal/model"
)

func CreateTransactionTx(tx *sql.Tx, t *model.Transaction) error {
	if t.NetAmount == 0 {
		t.NetAmount = t.Amount
	}

	query := `
    INSERT INTO transactions (
        user_id,
        date,
        amount,
        net_amount,
        type,
        is_transfer,
        place,
        note,
        method_id,
        category_id
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
    RETURNING transaction_id
    `

	log.Printf("📦 INSERT(TX): %+v", t)

	return tx.QueryRow(
		query,
		t.UserID,
		t.Date,
		t.Amount,
		t.NetAmount,
		t.Type,
		t.IsTransfer,
		t.Place,
		t.Note,
		t.MethodID,
		t.CategoryID,
	).Scan(&t.TransactionID)
}

func GetTransactions(db *sql.DB, userID int64) ([]model.Transaction, error) {
	query := `
    SELECT 
        transaction_id,
        user_id,
        date,
        amount,
        net_amount,
        type,
        is_transfer,
        place,
        note,
        method_id,
        category_id
    FROM transactions
    WHERE user_id = $1
    ORDER BY date DESC, transaction_id DESC
    `

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Transaction

	for rows.Next() {
		var t model.Transaction

		var place sql.NullString
		var note sql.NullString
		var categoryID sql.NullInt64

		err := rows.Scan(
			&t.TransactionID,
			&t.UserID,
			&t.Date,
			&t.Amount,
			&t.NetAmount,
			&t.Type,
			&t.IsTransfer,
			&place,
			&note,
			&t.MethodID,
			&categoryID,
		)
		if err != nil {
			return nil, err
		}

		if place.Valid {
			v := place.String
			t.Place = &v
		} else {
			t.Place = nil
		}

		if note.Valid {
			v := note.String
			t.Note = &v
		} else {
			t.Note = nil
		}

		if categoryID.Valid {
			v := categoryID.Int64
			t.CategoryID = &v
		} else {
			t.CategoryID = nil
		}

		list = append(list, t)
	}

	return list, rows.Err()
}
