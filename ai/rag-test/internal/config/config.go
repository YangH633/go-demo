package config

import (
	"os"
	"strconv"
)

type Config struct {
	OllamaURL      string
	EmbeddingURL   string
	QdrantURL      string
	QdrantAPIKey   string
	CollectionName string
	ChunkSize      int
	ChunkOverlap   int
	BatchSize      int // Number of chunks to process in each batch
}

func Load() *Config {
	return &Config{
		OllamaURL:      getEnv("OLLAMA_URL", "http://192.168.40.154:11434"),
		EmbeddingURL:   getEnv("EMBEDDING_URL", "http://192.168.40.154:11434/api/embeddings"),
		QdrantURL:      getEnv("QDRANT_URL", "http://192.168.40.154:6333"),
		QdrantAPIKey:   getEnv("QDRANT_API_KEY", ""),
		CollectionName: getEnv("COLLECTION_NAME", "rag_documents"),
		ChunkSize:      getEnvInt("CHUNK_SIZE", 500),
		ChunkOverlap:   getEnvInt("CHUNK_OVERLAP", 50),
		BatchSize:      getEnvInt("BATCH_SIZE", 20), // Process 20 chunks per batch to avoid memory issues (reduced from 50)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
