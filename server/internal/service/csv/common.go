package csv

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
)

// csvRow is one CSV row already normalized into transaction fields.
type csvRow struct {
	Date   time.Time
	Place  string
	Amount int
	Type   bool // true: income, false: expense
}

// csvRowParser turns a raw CSV record into a csvRow.
// ok=false, err=nil means the row should be silently skipped (header, short row, empty debit/credit, etc).
type csvRowParser func(record []string) (row csvRow, ok bool, err error)

// importCSVDir walks every *.csv file under dirPath, decodes it as Shift-JIS,
// parses each row with parseRow, reconciles it against manual transactions
// (repository.FindManualCandidatesForCsv), and imports the rows that don't match.
func importCSVDir(db *sql.DB, dirPath string, userID int, sourceLabel, methodIdentifier string, parseRow csvRowParser) error {
	fmt.Printf("📥 %s CSVのインポート開始\n", sourceLabel)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("ディレクトリ読み込み失敗: %w", err)
	}

	// メソッド情報を取得（全行で共通なので一度だけ取得）
	m, err := repository.GetMethodByIdentifier(db, methodIdentifier)
	if err != nil {
		return fmt.Errorf("メソッド取得エラー: %w", err)
	}

	totalCount := 0
	totalReconciledCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		fmt.Printf("📂 %s を処理中...\n", entry.Name())

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("ファイルオープン失敗: %w", err)
		}

		// ✅ Shift-JIS → UTF-8
		readerSJIS := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())

		reader := csv.NewReader(readerSJIS)
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1

		count := 0
		reconciledCount := 0

		for {
			record, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("⚠️ CSV読み込みエラー:", err)
				continue
			}

			row, ok, err := parseRow(record)
			if err != nil {
				fmt.Println("⚠️ 行パースエラー:", err)
				continue
			}
			if !ok {
				continue
			}

			place := row.Place
			t := &model.Transaction{
				TransactionID: 0,
				UserID:        int64(userID),
				Date:          row.Date,
				Type:          row.Type,
				Amount:        row.Amount,
				IsTransfer:    false,
				Place:         &place,
				MethodID:      int64(m.MethodID),
			}

			// ✅ 手入力トランザクションとの突合
			candidates, err := repository.FindManualCandidatesForCsv(db, t.UserID, t.MethodID, t.Date, t.Amount, place)
			if err != nil {
				fmt.Println("⚠️ 突合エラー:", err)
				continue
			}

			switch len(candidates) {
			case 1:
				if err := repository.MarkTransactionReconciled(db, candidates[0]); err != nil {
					fmt.Println("⚠️ 突合更新エラー:", err)
					continue
				}
				reconciledCount++
				totalReconciledCount++
				continue // CSV側は新規登録しない
			case 0:
				t.IsCsv = model.CsvStatusImported
			default:
				fmt.Printf("⚠️ 候補が複数件(%d件)見つかったためスキップ: place=%s amount=%d date=%s\n", len(candidates), place, t.Amount, row.Date.Format("2006/01/02"))
				continue
			}

			// ✅ place に応じて category_id / is_transfer を自動付与
			if err := repository.ApplyPlaceRule(db, t); err != nil {
				fmt.Println("⚠️ placeルール適用エラー:", err)
			}

			if err := repository.CreateTransaction(db, t); err != nil {
				fmt.Println("⚠️ DB登録失敗:", err)
				continue
			}

			count++
			totalCount++
		}

		file.Close()

		fmt.Printf("✅ %s から %d 件登録・%d 件突合確認\n", entry.Name(), count, reconciledCount)
	}

	fmt.Printf("🎉 合計 %d 件インポート・%d 件突合確認完了\n", totalCount, totalReconciledCount)
	return nil
}
