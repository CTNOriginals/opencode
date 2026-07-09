package neocortex

import (
	"sort"
	"strings"
	"unicode"
)

var stopWords = map[string]bool{
	"the": true, "a": true, "is": true, "and": true, "or": true,
	"of": true, "to": true, "in": true, "for": true, "on": true,
	"with": true, "at": true, "by": true, "an": true, "it": true,
	"as": true, "be": true, "this": true, "that": true, "was": true,
	"are": true, "were": true, "been": true, "has": true, "have": true,
	"had": true, "do": true, "does": true, "did": true, "but": true,
	"not": true, "so": true, "if": true, "no": true, "just": true,
	"about": true, "what": true, "which": true, "who": true, "how": true,
	"where": true, "when": true, "why": true,
}

func tokenize(text string) []string {
	var tokens []string
	var buf strings.Builder
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		} else {
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
		}
	}
	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}
	return tokens
}

func extractKeywords(text string) []string {
	return ExtractKeywords(text)
}

func ExtractKeywords(text string) []string {
	tokens := tokenize(text)
	seen := make(map[string]bool)
	var out []string
	for _, t := range tokens {
		if len(t) < 3 || stopWords[t] || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
		if len(out) >= 10 {
			break
		}
	}
	return out
}

func (s *Store) Search(query string, limit int) ([]MemoryItem, error) {
	words := tokenize(query)
	var keywords []string
	for _, w := range words {
		if !stopWords[w] && len(w) >= 3 {
			keywords = append(keywords, w)
		}
	}
	if len(keywords) > 5 {
		keywords = keywords[:5]
	}
	if len(keywords) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(keywords))
	args := make([]any, len(keywords))
	for i, k := range keywords {
		placeholders[i] = "?"
		args[i] = k
	}

	queryStr := `SELECT DISTINCT n.id, n.content, n.score, n.rate, n.created_at, n.updated_at
FROM neocortex n
JOIN neocortex_keywords nk ON nk.neocortex_id = n.id
WHERE nk.keyword IN (` + strings.Join(placeholders, ",") + `)
ORDER BY n.score DESC
LIMIT ?`

	args = append(args, limit)
	rows, err := s.reader.Query(queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MemoryItem
	for rows.Next() {
		var item MemoryItem
		if err := rows.Scan(&item.ID, &item.Content, &item.Score, &item.Rate, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return s.fallbackSearch(keywords, limit)
	}

	for i, item := range items {
		matchCount := 0
		itemLower := strings.ToLower(item.Content)
		for _, w := range keywords {
			if strings.Contains(itemLower, w) {
				matchCount++
			}
		}
		items[i].Score = item.Score * (1 + 0.1*float64(matchCount))
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})

	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *Store) fallbackSearch(words []string, limit int) ([]MemoryItem, error) {
	seen := make(map[int64]bool)
	var results []MemoryItem

	for _, w := range words {
		rows, err := s.reader.Query(
			`SELECT id, content, score, rate, created_at, updated_at FROM neocortex WHERE content LIKE ? ORDER BY score DESC LIMIT ?`,
			"%"+w+"%", limit,
		)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var item MemoryItem
			if err := rows.Scan(&item.ID, &item.Content, &item.Score, &item.Rate, &item.CreatedAt, &item.UpdatedAt); err != nil {
				rows.Close()
				return nil, err
			}
			if !seen[item.ID] {
				seen[item.ID] = true
				results = append(results, item)
			}
		}
		rows.Close()
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}
