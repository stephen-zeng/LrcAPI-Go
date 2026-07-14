package file

import (
	stderrors "errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

const currentLyricsSchemaVersion = 2

const lyricsMigrationKey = "lyrics_schema_version"

type lyricRow struct {
	CacheKey   string    `gorm:"column:cache_key;primaryKey;not null;index:idx_lyrics_cache_key"`
	LyricID    string    `gorm:"column:lyric_id;primaryKey;not null"`
	Title      string    `gorm:"column:title;not null"`
	Artist     string    `gorm:"column:artist;not null"`
	Lyric      string    `gorm:"column:lyric;not null"`
	Romaji     string    `gorm:"column:romaji;not null;default:''"`
	Type       string    `gorm:"column:type;not null;default:lrc"`
	Source     string    `gorm:"column:source;not null;default:''"`
	IsComplete int       `gorm:"column:is_complete;not null;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;default:CURRENT_TIMESTAMP"`
}

func (lyricRow) TableName() string {
	return "lyrics"
}

type schemaMigration struct {
	Key   string `gorm:"column:key;primaryKey;size:128"`
	Value string `gorm:"column:value;not null"`
}

func (schemaMigration) TableName() string {
	return "schema_migrations"
}

func migrateLyricsDB(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&lyricRow{}, &schemaMigration{}); err != nil {
			return pkgerrors.WithStack(err)
		}

		version, err := lyricsSchemaVersion(tx)
		if err != nil {
			return err
		}
		if version >= currentLyricsSchemaVersion {
			return nil
		}
		if version < 2 {
			if err := repairJoinedTranslationLines(tx); err != nil {
				return err
			}
		}
		if version < 1 {
			if err := backfillLyricsCompleteness(tx); err != nil {
				return err
			}
		}
		return saveLyricsSchemaVersion(tx, currentLyricsSchemaVersion)
	})
}

var migrationLrcTimeTagRe = regexp.MustCompile(`\[\d{1,2}:\d{2}[.:]\d{2,3}\]`)

// repairJoinedTranslationLines repairs the old blender output "[time]original[time]translation".
// Repeated tags without text between them are valid multi-timestamp LRC and remain unchanged.
func repairJoinedTranslationLines(db *gorm.DB) error {
	var rows []lyricRow
	if err := db.Find(&rows).Error; err != nil {
		return pkgerrors.WithStack(err)
	}
	for _, row := range rows {
		lyric, changed := splitJoinedTranslationLines(row.Lyric)
		if !changed {
			continue
		}
		if err := db.Model(&lyricRow{}).
			Where(lyricPrimaryKeyFilter(row.CacheKey, row.LyricID)).
			Update("lyric", lyric).Error; err != nil {
			return pkgerrors.WithStack(err)
		}
	}
	return nil
}

func splitJoinedTranslationLines(lrc string) (string, bool) {
	lines := strings.Split(lrc, "\n")
	changed := false
	for lineIndex, line := range lines {
		tags := migrationLrcTimeTagRe.FindAllStringIndex(line, -1)
		if len(tags) < 2 {
			continue
		}

		var repaired strings.Builder
		last := 0
		for i := 1; i < len(tags); i++ {
			previous, current := tags[i-1], tags[i]
			if line[previous[0]:previous[1]] != line[current[0]:current[1]] ||
				strings.TrimSpace(line[previous[1]:current[0]]) == "" {
				continue
			}
			repaired.WriteString(line[last:current[0]])
			repaired.WriteByte('\n')
			last = current[0]
			changed = true
		}
		if last > 0 {
			repaired.WriteString(line[last:])
			lines[lineIndex] = repaired.String()
		}
	}
	return strings.Join(lines, "\n"), changed
}

func lyricsSchemaVersion(db *gorm.DB) (int, error) {
	var migration schemaMigration
	err := db.First(&migration, &schemaMigration{Key: lyricsMigrationKey}).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, pkgerrors.WithStack(err)
	}
	version, err := strconv.Atoi(migration.Value)
	if err != nil {
		return 0, pkgerrors.WithStack(err)
	}
	return version, nil
}

func saveLyricsSchemaVersion(db *gorm.DB, version int) error {
	migration := schemaMigration{Key: lyricsMigrationKey, Value: strconv.Itoa(version)}
	if err := db.Save(&migration).Error; err != nil {
		return pkgerrors.WithStack(err)
	}
	return nil
}

func backfillLyricsCompleteness(db *gorm.DB) error {
	var rows []lyricRow
	if err := db.Find(&rows).Error; err != nil {
		return pkgerrors.WithStack(err)
	}
	for _, row := range rows {
		complete := boolToInt(computeComplete(row.Source, row.Lyric, row.Romaji))
		if complete == row.IsComplete {
			continue
		}
		if err := db.Model(&lyricRow{}).
			Where(lyricPrimaryKeyFilter(row.CacheKey, row.LyricID)).
			Select("IsComplete").
			Updates(lyricRow{IsComplete: complete}).Error; err != nil {
			return pkgerrors.WithStack(err)
		}
	}
	return nil
}

func lyricCacheKeyFilter(cacheKey string) map[string]interface{} {
	return map[string]interface{}{"cache_key": cacheKey}
}

func lyricPrimaryKeyFilter(cacheKey, lyricID string) map[string]interface{} {
	return map[string]interface{}{"cache_key": cacheKey, "lyric_id": lyricID}
}
