package repository

import (
	"database/sql"
	"errors"
	"log"
	"time"

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
        category_id,
        rule,
        is_csv
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
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
		t.Rule,
		t.IsCsv,
	).Scan(&t.TransactionID)
}

func CreateTransaction(db *sql.DB, t *model.Transaction) error {
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
        category_id,
        rule,
        is_csv
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
    RETURNING transaction_id
    `

	return db.QueryRow(
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
		t.Rule,
		t.IsCsv,
	).Scan(&t.TransactionID)
}

func UpdateTransaction(db *sql.DB, t *model.Transaction) error {
	if t.TransactionID <= 0 {
		return errors.New("transaction_id is required")
	}
	if t.NetAmount == 0 {
		t.NetAmount = t.Amount
	}

	query := `
    UPDATE transactions
    SET
        date = $1,
        amount = $2,
        net_amount = $3,
        type = $4,
        is_transfer = $5,
        place = $6,
        note = $7,
        method_id = $8,
        category_id = $9,
        rule = $10,
        is_csv = $11,
        updated_at = CURRENT_TIMESTAMP
    WHERE transaction_id = $12
    `

	result, err := db.Exec(
		query,
		t.Date,
		t.Amount,
		t.NetAmount,
		t.Type,
		t.IsTransfer,
		t.Place,
		t.Note,
		t.MethodID,
		t.CategoryID,
		t.Rule,
		t.IsCsv,
		t.TransactionID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func DeleteTransaction(db *sql.DB, transactionID int64) error {
	if transactionID <= 0 {
		return errors.New("transaction_id is required")
	}

	result, err := db.Exec(
		`DELETE FROM transactions WHERE transaction_id = $1`,
		transactionID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
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
        category_id,
        rule,
        is_csv
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
			&t.Rule,
			&t.IsCsv,
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

// FindManualCandidatesForCsv returns the transaction_ids of CsvStatusManual
// transactions matching the given user/method/date/amount and place (trim +
// case-insensitive exact match). These are candidates for CSV reconciliation.
func FindManualCandidatesForCsv(db *sql.DB, userID, methodID int64, date time.Time, amount int, place string) ([]int64, error) {
	query := `
    SELECT transaction_id
    FROM transactions
    WHERE user_id = $1
      AND method_id = $2
      AND date = $3
      AND amount = $4
      AND LOWER(TRIM(place)) = LOWER(TRIM($5))
      AND is_csv = $6
    `

	rows, err := db.Query(query, userID, methodID, date, amount, place, model.CsvStatusManual)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// MarkTransactionReconciled sets is_csv = CsvStatusReconciled for the given
// transaction, but only if it is currently CsvStatusManual.
func MarkTransactionReconciled(db *sql.DB, transactionID int64) error {
	if transactionID <= 0 {
		return errors.New("transaction_id is required")
	}

	result, err := db.Exec(
		`UPDATE transactions SET is_csv = $1, updated_at = CURRENT_TIMESTAMP WHERE transaction_id = $2 AND is_csv = $3`,
		model.CsvStatusReconciled,
		transactionID,
		model.CsvStatusManual,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
