# GHFS MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server for interacting with [GHFS (Go HTTP File Server)](https://github.com/mjpclab/go-http-file-server). Enables AI assistants to list directories, upload files, create directories, and delete files on a GHFS instance.

## Features

- **List directories** — browse GHFS directories with sorting support
- **Upload files/directories** — upload single files, multiple files, or entire directory structures
- **Create directories** — create one or more directories, including nested paths
- **Delete files/directories** — remove files or directories recursively
- **Archive/download** — generate archive download URLs (tar/tgz/zip)
- **Dual transport** — supports both STDIO and HTTP (Streamable HTTP) modes

## Build

```bash
go build -o ghfs-mcp-server .
```

## Usage

```
ghfs-mcp-server [options]

Options:
  -ghfs-url string   GHFS server base URL (default "http://localhost:8080")
  -mode string       Run mode: stdio or http (default "stdio")
  -addr string       HTTP listen address, only for http mode (default ":8080", or ":8443" with TLS)
  -cert string       TLS certificate file (enables HTTPS in http mode)
  -key string        TLS private key file (enables HTTPS in http mode)
  -debug             Enable debug logging of MCP messages
```

### STDIO Mode (default)

```bash
./ghfs-mcp-server -ghfs-url http://localhost:8080
```

### HTTP Mode

```bash
./ghfs-mcp-server -mode http -addr :9090 -ghfs-url http://localhost:8080
```

The HTTP endpoint is available at `http://localhost:9090/`.

### HTTPS Mode

```bash
./ghfs-mcp-server -mode http -addr :9443 -cert server.crt -key server.key -ghfs-url http://localhost:8080
```

The HTTPS endpoint is available at `https://localhost:9443/`.

## MCP Tools

### `ghfs_list`

List directory contents.

| Parameter | Type   | Required | Description                          |
| --------- | ------ | -------- | ------------------------------------ |
| `path`    | string | Yes      | Directory path, e.g. `/` or `/docs/` |
| `sort`    | string | No       | Sort order, e.g. `/T`, `n`, `S`      |

### `ghfs_upload`

Upload files or directory structures.

| Parameter          | Type     | Required | Description                                          |
| ------------------ | -------- | -------- | ---------------------------------------------------- |
| `path`             | string   | Yes      | Target directory path                                |
| `files`            | object[] | Yes      | Array of files to upload                             |
| `files[].filepath` | string   | Yes      | Relative path (e.g. `file.txt` or `subdir/file.txt`) |
| `files[].content`  | string   | Yes      | Base64-encoded file content                          |

Files with `/` in `filepath` are uploaded using GHFS `dirfile` mode, which auto-creates intermediate directories.

### `ghfs_mkdir`

Create directories.

| Parameter | Type     | Required | Description                                  |
| --------- | -------- | -------- | -------------------------------------------- |
| `path`    | string   | Yes      | Parent directory path                        |
| `names`   | string[] | Yes      | Directory names (supports nested like `a/b`) |

### `ghfs_delete`

Delete files or directories (recursive).

| Parameter | Type     | Required | Description              |
| --------- | -------- | -------- | ------------------------ |
| `path`    | string   | Yes      | Parent directory path    |
| `names`   | string[] | Yes      | Names of items to delete |

### `ghfs_archive`

Generate an archive download URL for files on the GHFS server.

| Parameter  | Type     | Required | Description                                     |
| ---------- | -------- | -------- | ----------------------------------------------- |
| `path`     | string   | Yes      | Directory path to archive                       |
| `format`   | string   | Yes      | Archive format: `tar`, `tgz`, or `zip`          |
| `names`    | string[] | No       | Specific items to include (omit for entire dir) |
| `filename` | string   | No       | Custom filename for the archive download        |

## Client Configuration

### Claude Desktop

STDIO mode — add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "ghfs": {
      "command": "/path/to/ghfs-mcp-server",
      "args": ["-ghfs-url", "http://localhost:8080"]
    }
  }
}
```

### VS Code (Copilot)

config file is `.vscode/mcp.json` (project scope) or `~/.config/Code/User/mcp.json` (user scope).

#### STDIO mode

```json
{
  "servers": {
    "ghfs": {
      "command": "/path/to/ghfs-mcp-server",
      "args": ["-ghfs-url", "http://localhost:8080"]
    }
  }
}
```

#### HTTP mode

```json
{
  "servers": {
    "ghfs": {
      "type": "http",
      "url": "http://localhost:9090/"
    }
  }
}
```
