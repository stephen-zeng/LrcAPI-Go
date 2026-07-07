package file

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type oldLyricRow struct {
	CacheKey  string    `gorm:"column:cache_key;primaryKey;not null;index:idx_lyrics_cache_key"`
	LyricID   string    `gorm:"column:lyric_id;primaryKey;not null"`
	Title     string    `gorm:"column:title;not null"`
	Artist    string    `gorm:"column:artist;not null"`
	Lyric     string    `gorm:"column:lyric;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;default:CURRENT_TIMESTAMP"`
}

func (oldLyricRow) TableName() string {
	return "lyrics"
}

func TestMigrateLyricsDBFromOldSchema(t *testing.T) {
	db := openTempGorm(t)
	defer closeDB(db)

	if err := db.AutoMigrate(&oldLyricRow{}); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	if err := db.Create(&[]oldLyricRow{
		{CacheKey: "zh", LyricID: "0", Title: "中文歌", Artist: "歌手", Lyric: "[00:01.00]你好世界"},
		{CacheKey: "ja", LyricID: "0", Title: "日本語", Artist: "歌手", Lyric: "[00:01.00]こんにちは"},
	}).Error; err != nil {
		t.Fatalf("insert old rows: %v", err)
	}
	if db.Migrator().HasColumn(&lyricRow{}, "Romaji") {
		t.Fatalf("old schema unexpectedly has romaji column")
	}

	if err := migrateLyricsDB(db); err != nil {
		t.Fatalf("migrateLyricsDB: %v", err)
	}

	for _, name := range []string{"romaji", "type", "source", "is_complete"} {
		if !db.Migrator().HasColumn(&lyricRow{}, name) {
			t.Fatalf("expected migrated column %q", name)
		}
	}

	var zh lyricRow
	if err := db.First(&zh, lyricPrimaryKeyFilter("zh", "0")).Error; err != nil {
		t.Fatalf("read migrated zh row: %v", err)
	}
	if zh.Romaji != "" || zh.Type != "lrc" || zh.Source != "" || zh.IsComplete != 1 {
		t.Fatalf("unexpected zh migration values: romaji=%q type=%q source=%q is_complete=%d", zh.Romaji, zh.Type, zh.Source, zh.IsComplete)
	}

	var ja lyricRow
	if err := db.First(&ja, lyricPrimaryKeyFilter("ja", "0")).Error; err != nil {
		t.Fatalf("read migrated ja row: %v", err)
	}
	if ja.IsComplete != 0 {
		t.Fatalf("expected incomplete japanese row without translation/romaji, got %d", ja.IsComplete)
	}

	version, err := lyricsSchemaVersion(db)
	if err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != currentLyricsSchemaVersion {
		t.Fatalf("expected schema version %d, got %d", currentLyricsSchemaVersion, version)
	}

	if err := migrateLyricsDB(db); err != nil {
		t.Fatalf("second migrateLyricsDB: %v", err)
	}
}

func TestWriteReadLyricRoundTrip(t *testing.T) {
	useTempLyricsDB(t)

	write := File{FolderName: "artist - title"}
	write.InfoLyric = []InfoLyric{
		{ID: "10", Title: "title", Artist: "artist", Lyric: "[00:10.00]后", Type: "lrc", Source: "netease"},
		{ID: "2", Title: "title", Artist: "artist", Lyric: "[00:02.00]前", Type: "lrc", Source: "qqmusic"},
	}
	if err := write.WriteLyric(); err != nil {
		t.Fatalf("WriteLyric: %v", err)
	}

	read := File{FolderName: write.FolderName}
	if err := read.ReadLyric(); err != nil {
		t.Fatalf("ReadLyric: %v", err)
	}
	if len(read.InfoLyric) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(read.InfoLyric))
	}
	if read.InfoLyric[0].ID != "2" || read.InfoLyric[1].ID != "10" {
		t.Fatalf("unexpected lyric order: %#v", read.InfoLyric)
	}
	for _, row := range read.InfoLyric {
		if !row.IsComplete {
			t.Fatalf("expected Chinese lyric %s to be complete", row.ID)
		}
	}
}

func openTempGorm(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "lyrics.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		t.Fatalf("ping sqlite: %v", err)
	}
	return db
}
