package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// Handler holds the GHFS client and provides MCP tool handler functions.
type Handler struct {
	client *ghfsClient
}

// NewHandler creates a new Handler with the given GHFS base URL.
func NewHandler(ghfsURL string) *Handler {
	return &Handler{client: newGHFSClient(ghfsURL)}
}

// HandleList handles the ghfs_list tool call.
func (h *Handler) HandleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: path"), nil
	}
	sort := req.GetString("sort", "")

	data, err := h.client.list(path, sort)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list directory: %v", err)), nil
	}

	// Pretty-print JSON for better readability
	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if pretty, err := json.MarshalIndent(raw, "", "  "); err == nil {
			data = pretty
		}
	}

	return mcp.NewToolResultText(string(data)), nil
}

// HandleUpload handles the ghfs_upload tool call.
func (h *Handler) HandleUpload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: path"), nil
	}

	args := req.GetArguments()
	filesRaw, ok := args["files"]
	if !ok {
		return mcp.NewToolResultError("missing required parameter: files"), nil
	}

	filesSlice, ok := filesRaw.([]any)
	if !ok {
		return mcp.NewToolResultError("parameter 'files' must be an array"), nil
	}

	if len(filesSlice) == 0 {
		return mcp.NewToolResultError("parameter 'files' must not be empty"), nil
	}

	var uploadFiles []uploadFile
	for i, item := range filesSlice {
		fileMap, ok := item.(map[string]any)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("files[%d]: must be an object with 'filepath' and 'content'", i)), nil
		}

		fp, ok := fileMap["filepath"].(string)
		if !ok || fp == "" {
			return mcp.NewToolResultError(fmt.Sprintf("files[%d]: missing or invalid 'filepath'", i)), nil
		}

		contentB64, ok := fileMap["content"].(string)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("files[%d]: missing or invalid 'content'", i)), nil
		}

		content, err := base64.StdEncoding.DecodeString(contentB64)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("files[%d]: invalid base64 content: %v", i, err)), nil
		}

		uploadFiles = append(uploadFiles, uploadFile{
			filepath: fp,
			content:  content,
		})
	}

	if err := h.client.upload(path, uploadFiles); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("upload failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully uploaded %d file(s) to %s", len(uploadFiles), path)), nil
}

// HandleMkdir handles the ghfs_mkdir tool call.
func (h *Handler) HandleMkdir(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: path"), nil
	}

	names := req.GetStringSlice("names", nil)
	if len(names) == 0 {
		return mcp.NewToolResultError("missing required parameter: names (must be a non-empty array of strings)"), nil
	}

	if err := h.client.mkdir(path, names); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("mkdir failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created %d directory(ies) under %s: %v", len(names), path, names)), nil
}

// HandleDelete handles the ghfs_delete tool call.
func (h *Handler) HandleDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: path"), nil
	}

	names := req.GetStringSlice("names", nil)
	if len(names) == 0 {
		return mcp.NewToolResultError("missing required parameter: names (must be a non-empty array of strings)"), nil
	}

	if err := h.client.delete(path, names); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("delete failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted %d item(s) from %s: %v", len(names), path, names)), nil
}

// HandleArchive handles the ghfs_archive tool call.
func (h *Handler) HandleArchive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: path"), nil
	}

	format, err := req.RequireString("format")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: format"), nil
	}

	switch format {
	case "tar", "tgz", "zip":
		// valid
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid format %q: must be \"tar\", \"tgz\", or \"zip\"", format)), nil
	}

	names := req.GetStringSlice("names", nil)
	filename := req.GetString("filename", "")

	downloadURL := h.client.archiveURL(path, format, names, filename)

	result := fmt.Sprintf("Archive download URL:\n%s\n\nFormat: %s", downloadURL, format)
	if len(names) > 0 {
		result += fmt.Sprintf("\nItems: %v", names)
	} else {
		result += fmt.Sprintf("\nScope: entire directory %s", path)
	}
	if filename != "" {
		result += fmt.Sprintf("\nFilename: %s", filename)
	}

	return mcp.NewToolResultText(result), nil
}
