package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type textEvent struct {
	Type string `json:"type"`
	Part struct {
		Text string `json:"text"`
	} `json:"part"`
}

// NewCaller returns an LLMFunc that calls opencode run via exec.Command.
// It uses the big-pickle model which reliably returns structured JSON.
func NewCaller() func(prompt string) (string, error) {
	return func(prompt string) (string, error) {
		cmd := exec.Command(
			"/home/ctn/.opencode/bin/opencode",
			"run", "--format", "json",
			"-m", "opencode/big-pickle",
			"--pure",
			prompt,
		)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return "", fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return "", fmt.Errorf("failed to start opencode: %w", err)
		}

		var textParts []string
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var event textEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				continue
			}
			if event.Type == "text" && event.Part.Text != "" {
				textParts = append(textParts, event.Part.Text)
			}
		}

		if err := cmd.Wait(); err != nil {
			log.Printf("opencode run failed: %v, stderr: %s", err, stderr.String())
			return "", fmt.Errorf("opencode run failed: %w", err)
		}

		raw := strings.Join(textParts, "")
		return extractJSON(raw), nil
	}
}

// extractJSON strips markdown code fences and extracts the first JSON object
// or array from the text, handling cases where the LLM adds extra commentary.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	// Strip markdown code fence if present
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	s = strings.TrimSpace(s)

	// Find the first { or [ to locate the JSON start
	start := -1
	for i, c := range s {
		if c == '{' || c == '[' {
			start = i
			break
		}
	}
	if start == -1 {
		return s
	}

	// Find the matching closing bracket
	depth := 0
	var open, close byte
	if s[start] == '{' {
		open, close = '{', '}'
	} else {
		open, close = '[', ']'
	}

	for i := start; i < len(s); i++ {
		if s[i] == open {
			depth++
		} else if s[i] == close {
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}

	return s[start:]
}
