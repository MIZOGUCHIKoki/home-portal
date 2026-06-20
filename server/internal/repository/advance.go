package repository

import (
	"database/sql"
	"errors"

	"kakeibo/internal/model"
)

func CreateAdvanceTx(tx *sql.Tx, transactionID int64, name string, amount int) error {
	if transactionID <= 0 {
		return errors.New("transaction_id is required")
	}
	if name == "" {
		return errors.New("name is required")
	}
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	_, err := tx.Exec(`
        INSERT INTO advance (
            transaction_id,
            name,
            amount,
            returned_amount,
            status
        )
        VALUES ($1, $2, $3, 0, FALSE)
    `,
		transactionID,
		name,
		amount,
	)

	return err
}

func ApplyRefundTx(tx *sql.Tx, advanceID int64, refundAmount int) error {
	if advanceID <= 0 {
		return errors.New("advance_id is required")
	}
	if refundAmount <= 0 {
		return errors.New("refund amount must be positive")
	}

	_, err := tx.Exec(`
        UPDATE advance
        SET
            returned_amount = returned_amount + $1,
            status = (returned_amount + $1) >= amount,
            updated_at = CURRENT_TIMESTAMP
        WHERE advance_id = $2
    `,
		refundAmount,
		advanceID,
	)

	return err
}

func GetAdvancesByTransactionID(db *sql.DB, transactionID int64) ([]model.Advance, error) {
	if transactionID <= 0 {
		return nil, errors.New("transaction_id is required")
	}

	rows, err := db.Query(`
        SELECT
            advance_id,
            transaction_id,
            name,
            amount,
            returned_amount,
            status,
            created_at,
            updated_at
        FROM advance
        WHERE transaction_id = $1
        ORDER BY advance_id
    `, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Advance

	for rows.Next() {
		var a model.Advance

		if err := rows.Scan(
			&a.AdvanceID,
			&a.TransactionID,
			&a.Name,
			&a.Amount,
			&a.ReturnedAmount,
			&a.Status,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, a)
	}

	return list, rows.Err()
}
