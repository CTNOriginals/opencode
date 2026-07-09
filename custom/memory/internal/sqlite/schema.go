package sqlite

import "database/sql"

func ApplySchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS hippocampus (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT,
			score REAL,
			weight REAL,
			rate REAL DEFAULT 0.95,
			tick_count INTEGER,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_hippocampus_score ON hippocampus(score)`,
		`CREATE INDEX IF NOT EXISTS idx_hippocampus_updated_at ON hippocampus(updated_at)`,
		`CREATE TABLE IF NOT EXISTS neocortex (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT,
			score REAL,
			rate REAL DEFAULT 0.98,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS neocortex_keywords (
			neocortex_id INTEGER NOT NULL,
			keyword TEXT NOT NULL,
			PRIMARY KEY (neocortex_id, keyword),
			FOREIGN KEY (neocortex_id) REFERENCES neocortex(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_neocortex_keywords_keyword ON neocortex_keywords(keyword)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
