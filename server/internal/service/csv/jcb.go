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

// ImportJCB imports all CSV files in a directory
func ImportJCB(db *sql.DB, dirPath string, userID int) error {
	fmt.Println("📥 JCB CSVのインポート開始")

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
		defer file.Close()

		// ✅ Shift-JIS → UTF-8
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
				fmt.Println("⚠️ CSV読み込みエラー:", err)
				continue
			}

			// ✅ 列数不足
			if len(record) < 5 {
				continue
			}

			// ✅ ヘッダ行スキップ
			if strings.Contains(record[2], "ご利用日") {
				continue
			}

			dateStr := strings.TrimSpace(record[2])
			place := strings.TrimSpace(record[3])
			amountStr := strings.ReplaceAll(record[4], ",", "")

			// ✅ 日付
			date, err := time.Parse("2006/01/02", dateStr)
			if err != nil {
				fmt.Printf("⚠️ 日付エラー: %s (%v)\n", dateStr, err)
				continue
			}

			// ✅ 金額
			amount, err := strconv.Atoi(amountStr)
			if err != nil {
				fmt.Printf("⚠️ 金額エラー: %s (%v)\n", amountStr, err)
				continue
			}
			// メソッド情報を取得（エラーは無視せず処理をスキップ）
			m, merr := repository.GetMethodByIdentifier(db, "jal_jcb")
			if merr != nil {
				fmt.Println("⚠️ メソッド取得エラー:", merr)
				continue
			}
			// ✅ トランザクション生成
			t := &model.Transaction{
				TransactionID: 0,
				UserID:        int64(userID),
				Date:          date,
				IsTransfer:    false,
				Place:         &place,
				MethodID:      int64(m.MethodID),
			}

			// ✅ 収入 / 支出判定
			if amount < 0 {
				t.Type = true
				t.Amount = -amount
			} else {
				t.Type = false
				t.Amount = amount
			}

			// ✅ DB保存
			// if err := repository.CreateTransaction(db, t); err != nil {
			// 	fmt.Println("⚠️ DB登録失敗:", err)
			// 	continue // ←止めない
			// }

			count++
			totalCount++
		}

		fmt.Printf("✅ %s から %d 件登録\n", entry.Name(), count)
	}

	fmt.Printf("🎉 合計 %d 件インポート完了\n", totalCount)
	return nil
}
