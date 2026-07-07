package file

import (
	"log"
	"lrcAPI/util"
	"sync"

	"github.com/pkg/errors"
)

// 同一 cache_key 的补全任务去重，避免并发请求重复调用大模型。
var completeInflight sync.Map

// computeComplete 计算一条歌词是否完备：fallback 占位恒完备，其余按内容判定。
func computeComplete(source, lyric, romaji string) bool {
	return source == "fallback" || util.IsLyricComplete(lyric, romaji)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// dbRow 是补全流程读出的单条歌词。
type dbRow struct {
	lyricID  string
	lyric    string
	romaji   string
	source   string
	complete bool
}

// CompleteLyricsAsync 是异步、无返回值的歌词补全入口。
//
// 它接收 title 与 artist，在后台协程中：
//  1. 按 `${artist} - ${title}` 在数据库中检索候选歌词；
//  2. 通过 IsLyricComplete 判断每条是否完备（外语缺中文翻译、日/韩缺罗马音即不完备）；
//  3. 对不完备的条目调用大模型补齐，并做行数一致性校验以防幻觉；
//  4. 将补齐后的结果写回数据库。
//
// 从客户端角度看该过程完全无感：调用后立即返回，补全在后台进行，
// 结果会在下一次查询同一歌曲时体现。
func CompleteLyricsAsync(title, artist string) {
	if !util.LLMEnabled() {
		return
	}
	cacheKey := artist + " - " + title
	// 已有同键任务在跑则跳过
	if _, loaded := completeInflight.LoadOrStore(cacheKey, struct{}{}); loaded {
		return
	}
	go func() {
		defer completeInflight.Delete(cacheKey)
		if err := completeLyrics(cacheKey); err != nil {
			util.ErrorPrinter(errors.Wrap(err, "completeLyrics"))
		}
	}()
}

// completeLyrics 执行实际的检索、补全与写回（同步执行，由协程调用）。
func completeLyrics(cacheKey string) error {
	rows, err := loadRows(cacheKey)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	// 收集需要更新的行；大模型调用不持有数据库连接。
	type update struct {
		lyricID  string
		lyric    string
		romaji   string
		complete bool
	}
	var updates []update

	for _, row := range rows {
		// fallback 占位与已标记完备的行无需处理（直接检查存储的 is_complete）
		if row.source == "fallback" || row.complete {
			continue
		}
		newLyric, newRomaji, changed, err := util.CompleteLyric(row.lyric, row.romaji)
		if err != nil {
			util.ErrorPrinter(errors.Wrap(err, "CompleteLyric "+cacheKey+"#"+row.lyricID))
			continue
		}
		nowComplete := computeComplete(row.source, newLyric, newRomaji)
		// 内容有更新，或存储的 is_complete 与实际不一致（自愈旧数据）时都写回
		if changed || nowComplete != row.complete {
			updates = append(updates, update{
				lyricID:  row.lyricID,
				lyric:    newLyric,
				romaji:   newRomaji,
				complete: nowComplete,
			})
		}
	}

	if len(updates) == 0 {
		return nil
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}
	stmt, err := tx.Prepare("UPDATE lyrics SET lyric = ?, romaji = ?, is_complete = ? WHERE cache_key = ? AND lyric_id = ?")
	if err != nil {
		_ = tx.Rollback()
		return errors.WithStack(err)
	}
	defer stmt.Close()
	for _, u := range updates {
		if _, err := stmt.Exec(u.lyric, u.romaji, boolToInt(u.complete), cacheKey, u.lyricID); err != nil {
			_ = tx.Rollback()
			return errors.WithStack(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	log.Printf("lyric completion: updated %d entries for %q", len(updates), cacheKey)
	return nil
}

// loadRows 读取某个 cache_key 下的全部歌词行后立即关闭连接。
func loadRows(cacheKey string) ([]dbRow, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT lyric_id, lyric, romaji, source, is_complete FROM lyrics WHERE cache_key = ?", cacheKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	var out []dbRow
	for rows.Next() {
		var r dbRow
		var complete int
		if err := rows.Scan(&r.lyricID, &r.lyric, &r.romaji, &r.source, &complete); err != nil {
			return nil, errors.WithStack(err)
		}
		r.complete = complete != 0
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}
