package server

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// ListTool defines the ghfs_list tool for listing directory contents.
var ListTool = mcp.NewTool("ghfs_list",
	mcp.WithDescription("List directory contents on the GHFS server. Returns JSON with file/directory names, sizes, modification times, and types."),
	mcp.WithString("path",
		mcp.Description("Directory path to list, e.g. \"/\" or \"/docs/\""),
		mcp.Required(),
	),
	mcp.WithString("sort",
		mcp.Description("Sort order for directory listing. "+
			"Key: \"n\" name, \"e\" extension/type, \"s\" size, \"t\" time, \"_\" no sort. "+
			"Uppercase reverses order (e.g. \"N\" name desc). "+
			"Prefix \"/\" puts dirs first (e.g. \"/n\"), suffix \"/\" puts dirs last (e.g. \"n/\"). "+
			"Examples: \"/T\" (dirs first, time desc), \"n\" (name asc), \"S\" (size desc)."),
	),
)

// UploadTool defines the ghfs_upload tool for uploading files/directories.
var UploadTool = mcp.NewTool("ghfs_upload",
	mcp.WithDescription("Upload one or more files to the GHFS server. Supports uploading directory structures by using relative paths with \"/\" in filepath. Files with \"/\" in filepath use \"dirfile\" mode which auto-creates subdirectories."),
	mcp.WithString("path",
		mcp.Description("Target directory path on GHFS, e.g. \"/ttt/\""),
		mcp.Required(),
	),
	mcp.WithArray("files",
		mcp.Description("Array of files to upload. Each file has \"filepath\" (relative path, e.g. \"file.txt\" or \"subdir/file.txt\") and \"content\" (base64-encoded file content)."),
		mcp.Required(),
		mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filepath": map[string]any{
					"type":        "string",
					"description": "Relative file path including filename. Use \"/\" for directory structure, e.g. \"subdir/file.txt\"",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Base64-encoded file content",
				},
			},
			"required": []string{"filepath", "content"},
		}),
	),
)

// MkdirTool defines the ghfs_mkdir tool for creating directories.
var MkdirTool = mcp.NewTool("ghfs_mkdir",
	mcp.WithDescription("Create one or more directories on the GHFS server. Supports creating nested directories like \"foo/bar/baz\"."),
	mcp.WithString("path",
		mcp.Description("Parent directory path, e.g. \"/ttt/\""),
		mcp.Required(),
	),
	mcp.WithArray("names",
		mcp.Description("Directory names to create. Supports nested paths like \"foo/bar\"."),
		mcp.Required(),
		mcp.WithStringItems(),
	),
)

// DeleteTool defines the ghfs_delete tool for deleting files/directories.
var DeleteTool = mcp.NewTool("ghfs_delete",
	mcp.WithDescription("Delete one or more files or directories on the GHFS server. Directories are deleted recursively."),
	mcp.WithString("path",
		mcp.Description("Parent directory path containing the items to delete, e.g. \"/ttt/\""),
		mcp.Required(),
	),
	mcp.WithArray("names",
		mcp.Description("Names of files or directories to delete."),
		mcp.Required(),
		mcp.WithStringItems(),
	),
)

// ArchiveTool defines the ghfs_archive tool for archiving files/directories.
var ArchiveTool = mcp.NewTool("ghfs_archive",
	mcp.WithDescription("Archive (download as package) files or directories on the GHFS server. Returns a download URL for the archive. Supports tar, tgz and zip formats."),
	mcp.WithString("path",
		mcp.Description("Directory path to archive, e.g. \"/ttt/mydir/\""),
		mcp.Required(),
	),
	mcp.WithString("format",
		mcp.Description("Archive format: \"tar\", \"tgz\", or \"zip\""),
		mcp.Required(),
		mcp.Enum("tar", "tgz", "zip"),
	),
	mcp.WithArray("names",
		mcp.Description("Optional: specific sub-item names to include in the archive. If omitted, the entire directory is archived."),
		mcp.WithStringItems(),
	),
	mcp.WithString("filename",
		mcp.Description("Optional: custom filename for the downloaded archive"),
	),
)
