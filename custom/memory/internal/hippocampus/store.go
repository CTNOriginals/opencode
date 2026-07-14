package hippocampus

import (
	"database/sql"
	"math"
	"time"
)

const (
	DefaultRate       = 0.95
	PromoteThreshold  = 0.7
	ForgetThreshold   = 0.3
)

type MemoryItem struct {
	ID        int64
	Content   string
	Score     float64
	Weight    float64
	Rate      float64
	TickCount int
	CreatedAt string
	UpdatedAt string
}

type Store struct {
	writer *sql.DB
	reader *sql.DB
}

func NewStore(writer, reader *sql.DB) *Store {
	return &Store{writer: writer, reader: reader}
}

func CalculateScore(weight, rate float64, tickCount int) float64 {
	return weight * math.Pow(rate, float64(tickCount))
}

func (s *Store) Insert(content string, score, weight, rate float64, tickCount int) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	res, err := s.writer.Exec(
		`INSERT INTO hippocampus (content, score, weight, rate, tick_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		content, score, weight, rate, tickCount, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) Get(id int64) (*MemoryItem, error) {
	row := s.reader.QueryRow(
		`SELECT id, content, score, weight, rate, tick_count, created_at, updated_at
		 FROM hippocampus WHERE id = ?`, id,
	)
	item := &MemoryItem{}
	err := row.Scan(&item.ID, &item.Content, &item.Score, &item.Weight, &item.Rate,
		&item.TickCount, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Store) ListAboveScore(threshold float64) ([]MemoryItem, error) {
	rows, err := s.reader.Query(
		`SELECT id, content, score, weight, rate, tick_count, created_at, updated_at
		 FROM hippocampus WHERE score > ?`, threshold,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MemoryItem
	for rows.Next() {
		var item MemoryItem
		err := rows.Scan(&item.ID, &item.Content, &item.Score, &item.Weight, &item.Rate,
			&item.TickCount, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) Delete(id int64) error {
	_, err := s.writer.Exec(`DELETE FROM hippocampus WHERE id = ?`, id)
	return err
}

func (s *Store) UpdateScore(id int64, score float64, tickCount int) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.writer.Exec(
		`UPDATE hippocampus SET score = ?, tick_count = ?, updated_at = ? WHERE id = ?`,
		score, tickCount, now, id,
	)
	return err
}

func (s *Store) AdvanceTicks() ([]MemoryItem, error) {
	// Read all items first, then close rows before issuing writes.
	// This avoids deadlock with MaxOpenConns=1 on the writer connection.
	rows, err := s.writer.Query(
		`SELECT id, content, score, weight, rate, tick_count, created_at, updated_at
		 FROM hippocampus`,
	)
	if err != nil {
		return nil, err
	}

	var items []MemoryItem
	for rows.Next() {
		var item MemoryItem
		err := rows.Scan(&item.ID, &item.Content, &item.Score, &item.Weight, &item.Rate,
			&item.TickCount, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	now := time.Now().Format(time.RFC3339)
	var promoted []MemoryItem

	for _, item := range items {
		item.TickCount++
		newScore := CalculateScore(item.Weight, item.Rate, item.TickCount)

		if newScore > PromoteThreshold {
			promoted = append(promoted, MemoryItem{
				ID:        item.ID,
				Content:   item.Content,
				Score:     newScore,
				Weight:    item.Weight,
				Rate:      item.Rate,
				TickCount: item.TickCount,
				CreatedAt: item.CreatedAt,
				UpdatedAt: now,
			})
		}

		if newScore < ForgetThreshold {
			_, err := s.writer.Exec(`DELETE FROM hippocampus WHERE id = ?`, item.ID)
			if err != nil {
				return nil, err
			}
			continue
		}

		_, err = s.writer.Exec(
			`UPDATE hippocampus SET score = ?, tick_count = ?, updated_at = ? WHERE id = ?`,
			newScore, item.TickCount, now, item.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	return promoted, nil
}
