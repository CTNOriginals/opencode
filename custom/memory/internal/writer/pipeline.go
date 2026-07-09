package writer

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/opencode/memory/internal/hippocampus"
	"github.com/opencode/memory/internal/neocortex"
	"github.com/opencode/memory/internal/pfc"
)

// LLMFunc abstracts the LLM call for testing with a stub.
type LLMFunc func(prompt string) (string, error)

type llmResponse struct {
	ReuseDecisions []reuseDecision `json:"reuse_decisions"`
	NewItems       []newItem       `json:"new_items"`
}

type reuseDecision struct {
	PFCIndex       int     `json:"pfc_index"`
	UpdatedContent string  `json:"updated_content"`
	NewWeight      float64 `json:"new_weight"`
}

type newItem struct {
	Content string  `json:"content"`
	Weight  float64 `json:"weight"`
	Rate    float64 `json:"rate"`
}

// Pipeline orchestrates the per-turn memory pipeline.
type Pipeline struct {
	rb  *pfc.RingBuffer
	hc  *hippocampus.Store
	nc  *neocortex.Store
	llm LLMFunc
}

func NewPipeline(rb *pfc.RingBuffer, hc *hippocampus.Store, nc *neocortex.Store, llm LLMFunc) *Pipeline {
	return &Pipeline{rb: rb, hc: hc, nc: nc, llm: llm}
}

func (p *Pipeline) buildPrompt(turnContext string) string {
	items := p.rb.Items()
	var b strings.Builder
	b.WriteString("You are a memory management system. Given the current short-term memory items and new context, decide which items to reuse and what new items to create.\n\n")
	b.WriteString("Current PFC items:\n")
	for i, item := range items {
		b.WriteString(fmt.Sprintf("  [%d] content=%q score=%.4f weight=%.4f rate=%.4f ticks=%d\n", i, item.Content, item.Score, item.Weight, item.Rate, item.TickCount))
	}
	b.WriteString(fmt.Sprintf("\nNew turn context: %s\n\n", turnContext))
	b.WriteString(`Respond with JSON:
{
  "reuse_decisions": [
    {"pfc_index": <int>, "updated_content": "<string>", "new_weight": <float>}
  ],
  "new_items": [
    {"content": "<string>", "weight": <float>, "rate": <float>}
  ]
}`)
	return b.String()
}

// ProcessTurn orchestrates the full per-turn pipeline. All errors are logged
// to stderr via log.Printf but never returned.
func (p *Pipeline) ProcessTurn(turnContext string) {
	// Step 1: Build LLM prompt and parse response
	prompt := p.buildPrompt(turnContext)
	respStr, err := p.llm(prompt)
	if err != nil {
		log.Printf("LLM call failed: %v", err)
	} else {
		var resp llmResponse
		if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
			log.Printf("failed to parse LLM response: %v", err)
		} else {
			// Step 2: Process reuse decisions
			for _, d := range resp.ReuseDecisions {
				p.rb.Reuse(d.PFCIndex, d.UpdatedContent, d.NewWeight)
			}

			// Step 3: Add new items
			for _, ni := range resp.NewItems {
				item := pfc.MemoryItem{
					Content:   ni.Content,
					Score:     ni.Weight,
					Weight:    ni.Weight,
					Rate:      ni.Rate,
					TickCount: 0,
					CreatedAt: time.Now(),
				}
				p.rb.PushFront(item)
			}
		}
	}

	// Step 4: Advance PFC ticks
	promoted := p.rb.AdvanceTick()

	// Step 5: Promote to hippocampus
	for _, item := range promoted {
		_, err := p.hc.Insert(item.Content, item.Weight, item.Weight, item.Rate, item.TickCount)
		if err != nil {
			log.Printf("failed to insert into hippocampus: %v", err)
		}
	}

	// Step 6: Advance hippocampus ticks and promote to neocortex
	hcPromoted, err := p.hc.AdvanceTicks()
	if err != nil {
		log.Printf("failed to advance hippocampus ticks: %v", err)
	} else {
		for _, item := range hcPromoted {
			keywords := LLMKeywords(p.llm, item.Content)
			_, err := p.nc.Insert(item.Content, item.Score, keywords)
			if err != nil {
				log.Printf("failed to insert into neocortex: %v", err)
			}
		}
	}

	// Step 7: Forget neocortex items below 0.6
	_, err = p.nc.DeleteBelowScore(0.6)
	if err != nil {
		log.Printf("failed to forget neocortex items: %v", err)
	}
}

// LLMKeywords extracts keywords using the LLM, falling back to statistical
// extraction on any error.
func LLMKeywords(llm LLMFunc, content string) []string {
	prompt := "Extract 3-5 keywords from this text that describe what it is about. Return as a JSON array of strings.\n\n" + content
	resp, err := llm(prompt)
	if err != nil {
		log.Printf("LLM keywords failed: %v, falling back to statistical extraction", err)
		return neocortex.ExtractKeywords(content)
	}

	var keywords []string
	if err := json.Unmarshal([]byte(resp), &keywords); err != nil {
		log.Printf("failed to parse LLM keywords response: %v, falling back to statistical extraction", err)
		return neocortex.ExtractKeywords(content)
	}

	return keywords
}
