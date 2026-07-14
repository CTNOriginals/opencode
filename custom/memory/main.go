package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/opencode/memory/internal/hippocampus"
	"github.com/opencode/memory/internal/llm"
	"github.com/opencode/memory/internal/neocortex"
	"github.com/opencode/memory/internal/pfc"
	"github.com/opencode/memory/internal/sqlite"
	"github.com/opencode/memory/internal/writer"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("[memory-server] ")

	if len(os.Args) > 1 && os.Args[1] == "process-session" {
		runProcessSession()
		return
	}

	runMCPServer()
}

func runProcessSession() {
	dbPath := os.Getenv("MEMORY_DB_PATH")
	if dbPath == "" {
		dbPath = "memory.db"
	}

	writerDB, err := sqlite.OpenWriter(dbPath)
	if err != nil {
		log.Printf("failed to open writer connection: %v", err)
		return
	}
	defer writerDB.Close()

	readerDB, err := sqlite.OpenReader(dbPath)
	if err != nil {
		log.Printf("failed to open reader connection: %v", err)
		return
	}
	defer readerDB.Close()

	if err := sqlite.ApplySchema(writerDB); err != nil {
		log.Printf("failed to apply schema: %v", err)
		return
	}

	contextBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Printf("failed to read stdin: %v", err)
		return
	}
	turnContext := strings.TrimSpace(string(contextBytes))
	if turnContext == "" {
		log.Printf("empty turn context, nothing to process")
		return
	}

	hcStore := hippocampus.NewStore(writerDB, readerDB)
	ncStore := neocortex.NewStore(writerDB, readerDB)
	rb := pfc.NewRingBuffer(50)
	caller := llm.NewCaller()

	pipeline := writer.NewPipeline(rb, hcStore, ncStore, caller)
	pipeline.ProcessTurn(turnContext)

	log.Printf("session processed successfully")
}

func runMCPServer() {
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
	caller := llm.NewCaller()

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
			keywords = writer.LLMKeywords(caller, content)
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
