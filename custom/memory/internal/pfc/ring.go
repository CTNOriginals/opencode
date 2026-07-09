package pfc

import (
	"math"
	"time"
)

type MemoryItem struct {
	Content   string
	Score     float64
	Weight    float64
	Rate      float64
	TickCount int
	CreatedAt time.Time
}

func CalculateScore(weight, rate float64, tickCount int) float64 {
	return weight * math.Pow(rate, float64(tickCount))
}

type RingBuffer struct {
	items []MemoryItem
	size  int
	rate  float64
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		items: make([]MemoryItem, 0, size),
		size:  size,
		rate:  0.85,
	}
}

func (rb *RingBuffer) PushFront(item MemoryItem) {
	if len(rb.items) >= rb.size {
		rb.items = rb.items[:len(rb.items)-1]
	}
	rb.items = append([]MemoryItem{item}, rb.items...)
}

func (rb *RingBuffer) Reuse(index int, newContent string, llmWeight float64) {
	if index < 0 || index >= len(rb.items) {
		return
	}
	item := rb.items[index]
	item.Content = newContent
	item.Weight = 0.7*item.Weight + 0.3*llmWeight
	item.TickCount = 0
	item.Score = CalculateScore(item.Weight, rb.rate, item.TickCount)
	rb.items = append(rb.items[:index], rb.items[index+1:]...)
	rb.items = append([]MemoryItem{item}, rb.items...)
}

func (rb *RingBuffer) AdvanceTick() (promoted []MemoryItem) {
	for i := range rb.items {
		rb.items[i].TickCount++
		rb.items[i].Score = CalculateScore(rb.items[i].Weight, rb.rate, rb.items[i].TickCount)
	}
	for _, item := range rb.items {
		if item.Score > 0.7 {
			promoted = append(promoted, item)
		}
	}
	filtered := make([]MemoryItem, 0, len(rb.items))
	for _, item := range rb.items {
		if item.Score >= 0.5 {
			filtered = append(filtered, item)
		}
	}
	rb.items = filtered
	return promoted
}

func (rb *RingBuffer) Len() int {
	return len(rb.items)
}

func (rb *RingBuffer) Items() []MemoryItem {
	result := make([]MemoryItem, len(rb.items))
	copy(result, rb.items)
	return result
}

func (rb *RingBuffer) GetByIndex(index int) (MemoryItem, bool) {
	if index < 0 || index >= len(rb.items) {
		return MemoryItem{}, false
	}
	return rb.items[index], true
}
