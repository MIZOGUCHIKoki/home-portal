package csv

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
)

// rowParserFor resolves the csvRowParser for a known method identifier
// (mirrors the ImportJCB / ImportSBINetBank wiring in jcb.go / netbk.go).
func rowParserFor(methodIdentifier string) (csvRowParser, error) {
	switch methodIdentifier {
	case "jal_jcb":
		return parseJCBRow, nil
	case "0038":
		return parseNetBKRow, nil
	default:
		return nil, fmt.Errorf("unsupported method_identifier: %s", methodIdentifier)
	}
}

// CsvMismatch pairs a parsed CSV row with a manual transaction that appears
// to refer to the same real-world transaction (same method, and at least two
// of date/amount/place match) but disagrees on date, place, and/or amount.
type CsvMismatch struct {
	CsvDate     time.Time
	CsvPlace    string
	CsvAmount   int
	CsvType     bool
	Transaction model.Transaction
}

// CheckResult summarizes a dry-run comparison of a CSV file against the
// user's manually-entered (is_csv = CsvStatusManual) transactions.
type CheckResult struct {
	TotalRows int

	// ReconciledCount: rows whose date/place/amount exactly match a manual
	// transaction. These are treated as already-consistent and are not
	// reported (they are the same rows importCSVDir would auto-reconcile).
	ReconciledCount int

	// NewCount: rows with no manual transaction sharing date/amount/place
	// even partially. These are candidates for a fresh import, not a
	// consistency problem.
	NewCount int

	// Mismatches: rows that partially match a manual transaction (2 of 3
	// fields) but disagree on the rest — the inconsistencies the caller
	// asked to review.
	Mismatches []CsvMismatch
}

// CheckCsvAgainstManual parses CSV bytes (Shift-JIS encoded, as produced by
// the bank/card issuers this app imports from) using the parser registered
// for methodIdentifier, then compares every row against the user's manual
// transactions for that method. It performs no DB writes: transactions
// already reconciled (CsvStatusReconciled) are never touched or considered,
// since only CsvStatusManual rows are queried.
func CheckCsvAgainstManual(db *sql.DB, data []byte, userID int64, methodIdentifier string) (*CheckResult, error) {
	parseRow, err := rowParserFor(methodIdentifier)
	if err != nil {
		return nil, err
	}

	m, err := repository.GetMethodByIdentifier(db, methodIdentifier)
	if err != nil {
		return nil, fmt.Errorf("メソッド取得エラー: %w", err)
	}

	readerSJIS := transform.NewReader(bytes.NewReader(data), japanese.ShiftJIS.NewDecoder())

	reader := csv.NewReader(readerSJIS)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	result := &CheckResult{}
	seenMismatch := make(map[int64]bool)

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("CSV読み込みエラー: %w", err)
		}

		row, ok, err := parseRow(record)
		if err != nil {
			return nil, fmt.Errorf("行パースエラー: %w", err)
		}
		if !ok {
			continue
		}

		result.TotalRows++

		exact, err := repository.FindManualCandidatesForCsv(db, userID, int64(m.MethodID), row.Date, row.Amount, row.Place)
		if err != nil {
			return nil, fmt.Errorf("突合エラー: %w", err)
		}
		if len(exact) >= 1 {
			// 日付・場所・金額が完全一致 = 突合済みと同等なので報告不要
			result.ReconciledCount++
			continue
		}

		mismatchIDs, err := repository.FindManualMismatchCandidates(db, userID, int64(m.MethodID), row.Date, row.Amount, row.Place)
		if err != nil {
			return nil, fmt.Errorf("不整合チェックエラー: %w", err)
		}
		if len(mismatchIDs) == 0 {
			result.NewCount++
			continue
		}

		for _, id := range mismatchIDs {
			if seenMismatch[id] {
				continue
			}
			seenMismatch[id] = true

			t, err := repository.GetTransactionByID(db, id)
			if err != nil {
				return nil, fmt.Errorf("トランザクション取得エラー: %w", err)
			}

			result.Mismatches = append(result.Mismatches, CsvMismatch{
				CsvDate:     row.Date,
				CsvPlace:    row.Place,
				CsvAmount:   row.Amount,
				CsvType:     row.Type,
				Transaction: *t,
			})
		}
	}

	return result, nil
}
