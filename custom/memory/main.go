package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/opencode/memory/internal/neocortex"
	"github.com/opencode/memory/internal/sqlite"
	"github.com/opencode/memory/internal/writer"
)

var llmStub writer.LLMFunc = func(prompt string) (string, error) {
	return `{"reuse_decisions":[],"new_items":[]}`, nil
}

func main() {
	dbPath := os.Getenv("MEMORY_DB_PATH")
	if dbPath == "" {
		dbPath = "memory.db"
	}

	writerDB, err := sqlite.OpenWriter(dbPath)
	if err != nil {
		log.Fatalf("failed to open writer connection: %v", err)
	}
	defer writerDB.Close()

	readerDB, err := sqlite.OpenReader(dbPath)
	if err != nil {
		log.Fatalf("failed to open reader connection: %v", err)
	}
	defer readerDB.Close()

	if err := sqlite.ApplySchema(writerDB); err != nil {
		log.Fatalf("failed to apply schema: %v", err)
	}

	nc := neocortex.NewStore(writerDB, readerDB)

	s := server.NewMCPServer("memory-server", "1.0.0")

	readTool := mcp.NewTool("read_memory",
		mcp.WithDescription("Search stored memories by query"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query"),
		),
	)

	s.AddTool(readTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		query, ok := args["query"].(string)
		if !ok || query == "" {
			return mcp.NewToolResultError("query parameter is required"), nil
		}

		items, err := nc.Search(query, 10)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		if len(items) == 0 {
			return mcp.NewToolResultText("No results found."), nil
		}

		var b strings.Builder
		for i, item := range items {
			b.WriteString(fmt.Sprintf("%d. [id=%d] score=%.2f\n", i+1, item.ID, item.Score))
			b.WriteString(fmt.Sprintf("   Content: %s\n", item.Content))
			if len(item.Keywords) > 0 {
				b.WriteString(fmt.Sprintf("   Keywords: %s\n", strings.Join(item.Keywords, ", ")))
			}
			b.WriteString(fmt.Sprintf("   Created: %s\n", item.CreatedAt))
		}

		return mcp.NewToolResultText(b.String()), nil
	})

	writeTool := mcp.NewTool("write_memory",
		mcp.WithDescription("Store a new memory with optional keywords"),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The content to remember"),
		),
		mcp.WithString("keywords",
			mcp.Description("Comma-separated keywords"),
		),
	)

	s.AddTool(writeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		content, ok := args["content"].(string)
		if !ok || content == "" {
			return mcp.NewToolResultError("content parameter is required"), nil
		}

		var keywords []string
		if kwArg, ok := args["keywords"].(string); ok && kwArg != "" {
			for _, kw := range strings.Split(kwArg, ",") {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					keywords = append(keywords, kw)
				}
			}
		} else {
			keywords = writer.LLMKeywords(llmStub, content)
		}

		id, err := nc.Insert(content, 1.0, keywords)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to store memory: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Memory stored with ID: %d", id)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
