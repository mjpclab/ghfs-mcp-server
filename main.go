package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	mcpserver "ghfs-mcp-server/server"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	ghfsURL := flag.String("ghfs-url", "http://localhost:8080", "GHFS server base URL")
	mode := flag.String("mode", "stdio", "Run mode: stdio or http")
	certFile := flag.String("cert", "", "TLS certificate file (enables HTTPS in http mode)")
	keyFile := flag.String("key", "", "TLS private key file (enables HTTPS in http mode)")
	addr := flag.String("addr", "", "HTTP listen address, only for http mode (default \":8080\", or \":8443\" with TLS)")
	debug := flag.Bool("debug", false, "Enable debug logging of MCP messages")
	flag.Parse()

	if *addr == "" {
		if *certFile == "" || *keyFile == "" {
			*addr = ":8080"
		} else {
			*addr = ":8443"
		}
	}

	s := mcpserver.NewServer(*ghfsURL, *debug)

	switch *mode {
	case "stdio":
		log.Printf("Starting %s %s in STDIO mode (GHFS: %s)\n", mcpserver.ServerName, mcpserver.ServerVersion, *ghfsURL)
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("STDIO server error: %v", err)
		}

	case "http":
		opts := []server.StreamableHTTPOption{
			server.WithEndpointPath("/"),
		}
		if *certFile != "" && *keyFile != "" {
			opts = append(opts, server.WithTLSCert(*certFile, *keyFile))
			log.Println("Using HTTPS scheme")
		}
		log.Printf("Starting %s %s in HTTP mode on %s (GHFS: %s)\n", mcpserver.ServerName, mcpserver.ServerVersion, *addr, *ghfsURL)
		httpServer := server.NewStreamableHTTPServer(s, opts...)
		if err := httpServer.Start(*addr); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s (expected 'stdio' or 'http')\n", *mode)
		os.Exit(1)
	}
}
