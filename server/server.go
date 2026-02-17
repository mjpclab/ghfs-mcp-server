package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ServerName    = "ghfs-mcp-server"
	ServerVersion = "0.1.0"
)

// NewServer creates a new MCP server with all GHFS tools registered.
func NewServer(ghfsURL string, debug bool) *server.MCPServer {
	h := NewHandler(ghfsURL)

	opts := []server.ServerOption{
		server.WithToolCapabilities(true),
		server.WithInstructions(fmt.Sprintf(
			"This MCP server provides tools to interact with a GHFS (Go HTTP File Server) instance at %s. "+
				"Available operations: list directories, upload files/directories, create directories, delete files/directories.",
			ghfsURL,
		)),
	}

	if debug {
		opts = append(opts, server.WithHooks(newDebugHooks()))
	}

	s := server.NewMCPServer(ServerName, ServerVersion, opts...)

	s.AddTool(ListTool, h.HandleList)
	s.AddTool(UploadTool, h.HandleUpload)
	s.AddTool(MkdirTool, h.HandleMkdir)
	s.AddTool(DeleteTool, h.HandleDelete)
	s.AddTool(ArchiveTool, h.HandleArchive)

	return s
}

func newDebugHooks() *server.Hooks {
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		data, _ := json.MarshalIndent(message, "", "  ")
		log.Printf("[DEBUG] --> %s id=%v\n%s", method, id, data)
	})

	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		data, _ := json.MarshalIndent(result, "", "  ")
		log.Printf("[DEBUG] <-- %s id=%v\n%s", method, id, data)
	})

	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("[DEBUG] <-- %s id=%v ERROR: %v", method, id, err)
	})

	return hooks
}
