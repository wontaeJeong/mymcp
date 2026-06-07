package main

import (
	"fmt"
	"log"
	"net/http"

	"mymcp/internal/config"
	"mymcp/internal/server"
)

func main() {
	cfg := config.Load(nil)
	address := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("MCP OAuth demo listening on %s%s", cfg.BaseURL, cfg.MCPPath)
	log.Fatal(http.ListenAndServe(address, server.NewHandler(cfg, nil)))
}
