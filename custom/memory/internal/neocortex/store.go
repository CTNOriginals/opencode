package neocortex

import (
	"database/sql"
	"time"
)

type Store struct {
	writer *sql.DB
	reader *sql.DB
}

func NewStore(writer, reader *sql.DB) *Store {
	return &Store{writer: writer, reader: reader}
}

type MemoryItem struct {
	ID        int64
	Content   string
	Score     float64
	Rate      float64
	Keywords  []string
	CreatedAt string
	UpdatedAt string
}

func (s *Store) Insert(content string, score float64, keywords []string) (int64, error) {
	tx, err := s.writer.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now().Format(time.RFC3339)
	res, err := tx.Exec(
		`INSERT INTO neocortex (content, score, rate, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		content, score, 0.98, now, now,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, kw := range keywords {
		_, err := tx.Exec(
			`INSERT INTO neocortex_keywords (neocortex_id, keyword) VALUES (?, ?)`,
			id, kw,
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) Get(id int64) (*MemoryItem, error) {
	row := s.reader.QueryRow(
		`SELECT id, content, score, rate, created_at, updated_at FROM neocortex WHERE id = ?`, id,
	)
	item := &MemoryItem{}
	if err := row.Scan(&item.ID, &item.Content, &item.Score, &item.Rate, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	rows, err := s.reader.Query(
		`SELECT keyword FROM neocortex_keywords WHERE neocortex_id = ?`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var kw string
		if err := rows.Scan(&kw); err != nil {
			return nil, err
		}
		item.Keywords = append(item.Keywords, kw)
	}
	return item, rows.Err()
}

func (s *Store) Delete(id int64) error {
	_, err := s.writer.Exec(`DELETE FROM neocortex WHERE id = ?`, id)
	return err
}

func (s *Store) UpdateScore(id int64, score float64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.writer.Exec(
		`UPDATE neocortex SET score = ?, updated_at = ? WHERE id = ?`,
		score, now, id,
	)
	return err
}

func (s *Store) DeleteBelowScore(threshold float64) (int, error) {
	res, err := s.writer.Exec(`DELETE FROM neocortex WHERE score < ?`, threshold)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}
