package csv

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
)

// ImportSBINetBank imports ALL CSV files in a directory
func ImportSBINetBank(db *sql.DB, dirPath string, userID int) error {

	fmt.Println("📥 SBI銀行CSVのインポート開始")

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("ディレクトリ読み込み失敗: %w", err)
	}

	totalCount := 0

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

		// ✅ Shift-JIS対応
		readerSJIS := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())

		reader := csv.NewReader(readerSJIS)
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1

		count := 0

		for {
			record, err := reader.Read()

			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("⚠️ CSVエラー:", err)
				continue
			}

			// ✅ 列数チェック
			if len(record) < 5 {
				continue
			}

			// ✅ ヘッダスキップ
			if strings.Contains(record[0], "日付") {
				continue
			}

			dateStr := strings.TrimSpace(record[0])
			content := strings.TrimSpace(record[1])

			debitStr := strings.ReplaceAll(record[2], ",", "")
			creditStr := strings.ReplaceAll(record[3], ",", "")

			// ✅ 日付
			date, err := time.Parse("2006/01/02", dateStr)
			if err != nil {
				fmt.Println("⚠️ 日付エラー:", dateStr)
				continue
			}

			// メソッド情報を取得（エラーは無視せず処理をスキップ）
			m, merr := repository.GetMethodByIdentifier(db, "0038") // ドコモSMTBネット銀行
			if merr != nil {
				fmt.Println("⚠️ メソッド取得エラー:", merr)
				continue
			}

			t := &model.Transaction{
				TransactionID: 0,
				UserID:        int64(userID),
				Date:          date,
				IsTransfer:    false,
				Place:         &content,
				MethodID:      int64(m.MethodID),
			}

			// ✅ 出金（支出）
			if debitStr != "" {
				amount, err := strconv.Atoi(debitStr)
				if err != nil {
					fmt.Println("⚠️ 出金パースエラー:", debitStr)
					continue
				}
				t.Type = false
				t.Amount = amount

				// ✅ 入金（収入）
			} else if creditStr != "" {
				amount, err := strconv.Atoi(creditStr)
				if err != nil {
					fmt.Println("⚠️ 入金パースエラー:", creditStr)
					continue
				}
				t.Type = true
				t.Amount = amount
			} else {
				continue
			}

			// ✅ DB保存
			// if err := repository.CreateTransaction(db, t); err != nil {
			// 	fmt.Println("⚠️ DBエラー:", err)
			// 	continue
			// }

			count++
			totalCount++
		}

		file.Close()

		fmt.Printf("✅ %s から %d 件登録\n", entry.Name(), count)
	}

	fmt.Printf("🎉 合計 %d 件インポート完了\n", totalCount)
	return nil
}
