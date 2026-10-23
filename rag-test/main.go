package main

import (
	"flag"
	"log"
	"net/http"

	"rag-test/internal/api"
	"rag-test/internal/config"
	"rag-test/internal/rag"
	"rag-test/internal/storage"
)

func main() {
	port := flag.String("port", "8080", "Server port")
	flag.Parse()

	cfg := config.Load()

	// Initialize RAG service
	ragService := rag.NewRAGService(cfg)

	// Initialize storage
	storageService := storage.NewStorageService(cfg)

	// Setup API routes
	handler := api.NewHandler(ragService, storageService)

	log.Printf("Starting RAG server on port %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}
