package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	config := LoadConfig(nil)
	address := fmt.Sprintf(":%d", config.Port)
	log.Printf("MCP OAuth demo listening on %s%s", config.BaseURL, config.MCPPath)
	log.Fatal(http.ListenAndServe(address, CreateHandler(config, nil)))
}
