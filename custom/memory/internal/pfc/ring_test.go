package pfc

import (
	"fmt"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer(5)

	// 1. Create items at various weights/rates
	rb.PushFront(MemoryItem{Content: "low", Weight: 0.6, Rate: 0.85, TickCount: 0, Score: 0.6})
	rb.PushFront(MemoryItem{Content: "mid", Weight: 0.8, Rate: 0.85, TickCount: 0, Score: 0.8})
	rb.PushFront(MemoryItem{Content: "high", Weight: 1.0, Rate: 0.85, TickCount: 0, Score: 1.0})

	fmt.Println("=== Initial items ===")
	for _, item := range rb.Items() {
		fmt.Printf("  %s: weight=%.2f rate=%.2f ticks=%d score=%.4f\n", item.Content, item.Weight, item.Rate, item.TickCount, item.Score)
	}

	// 2. Advance ticks and check score decay
	promoted := rb.AdvanceTick()
	fmt.Println("\n=== After first tick ===")
	for _, item := range rb.Items() {
		fmt.Printf("  %s: score=%.4f\n", item.Content, item.Score)
	}

	// 3. Confirm eviction at < 0.5 and promotion at > 0.7
	fmt.Println("\n=== Promoted items (score > 0.7) ===")
	for _, p := range promoted {
		fmt.Printf("  %s: score=%.4f\n", p.Content, p.Score)
	}

	// Calculate expected scores after 1 tick
	// low: 0.6 * 0.85^1 = 0.51 -> kept, not promoted
	// mid: 0.8 * 0.85^1 = 0.68 -> kept, not promoted
	// high: 1.0 * 0.85^1 = 0.85 -> kept, promoted
	expectedLow := 0.6 * 0.85
	expectedMid := 0.8 * 0.85
	expectedHigh := 1.0 * 0.85

	items := rb.Items()
	if len(items) != 3 {
		t.Errorf("expected 3 items after first tick, got %d", len(items))
	}
	if items[0].Score != expectedHigh || items[1].Score != expectedMid || items[2].Score != expectedLow {
		t.Errorf("score mismatch: got %.4f, %.4f, %.4f", items[0].Score, items[1].Score, items[2].Score)
	}
	if len(promoted) != 1 || promoted[0].Content != "high" {
		t.Errorf("expected only 'high' promoted, got %+v", promoted)
	}

	// Tick again: should evict low (score < 0.5)
	promoted = rb.AdvanceTick()
	expectedLow = 0.6 * 0.85 * 0.85 // 0.4335
	fmt.Printf("\n=== After second tick (low should be < 0.5: %.4f) ===\n", expectedLow)
	for _, item := range rb.Items() {
		fmt.Printf("  %s: score=%.4f\n", item.Content, item.Score)
	}
	if len(rb.Items()) != 2 {
		t.Errorf("expected 2 items after eviction, got %d", len(rb.Items()))
	}

	// 4. Confirm weight blending on reuse (70/30)
	fmt.Println("\n=== Testing weight blending on reuse ===")
	rb2 := NewRingBuffer(5)
	rb2.PushFront(MemoryItem{Content: "original", Weight: 1.0, Rate: 0.85, TickCount: 0, Score: 1.0})
	fmt.Printf("Before reuse: weight=%.2f\n", rb2.Items()[0].Weight)
	rb2.Reuse(0, "reused", 0.5)
	expectedWeight := 0.7*1.0 + 0.3*0.5 // 0.85
	item, ok := rb2.GetByIndex(0)
	if !ok {
		t.Fatal("expected item after reuse")
	}
	fmt.Printf("After reuse: weight=%.2f (expected %.2f)\n", item.Weight, expectedWeight)
	if item.Weight != expectedWeight {
		t.Errorf("weight blending failed: got %.4f, want %.4f", item.Weight, expectedWeight)
	}
	if item.Content != "reused" {
		t.Errorf("content not updated: got %q, want 'reused'", item.Content)
	}
	if item.TickCount != 0 {
		t.Errorf("tick count not reset: got %d, want 0", item.TickCount)
	}

	fmt.Println("\n=== All checks passed ===")
}
