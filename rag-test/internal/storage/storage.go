package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"rag-test/internal/config"
	"rag-test/internal/parser"
)

type StorageService struct {
	cfg *config.Config
}

func NewStorageService(cfg *config.Config) *StorageService {
	return &StorageService{cfg: cfg}
}

// UploadDocument handles document upload and processing
// Uses batch processing to avoid memory issues with large documents
func (s *StorageService) UploadDocument(ctx context.Context, filename string, content []byte) error {
	// Parse document based on file type
	fmt.Printf("[处理中] 开始解析文档: %s\n", filename)
	text, err := s.parseDocument(filename, content)
	if err != nil {
		return fmt.Errorf("failed to parse document: %w", err)
	}
	fmt.Printf("[处理中] 文档解析完成，文本长度: %d 字符\n", len(text))

	// Clear content immediately after parsing to free memory
	content = nil

	// Split text into chunks
	fmt.Printf("[处理中] 开始文本分块...\n")
	chunks := s.splitText(text)
	fmt.Printf("[处理中] 文档已解析，共生成 %d 个文本块\n", len(chunks))

	// Clear text after splitting to free memory
	text = ""

	// Process in batches to avoid memory issues
	batchSize := s.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}
	totalBatches := (len(chunks) + batchSize - 1) / batchSize

	fmt.Printf("[处理中] 开始分批处理（每批 %d 个块，共 %d 批）...\n", batchSize, totalBatches)

	// Ensure collection exists before processing
	if err := s.ensureCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	for batchIdx := 0; batchIdx < totalBatches; batchIdx++ {
		start := batchIdx * batchSize
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		// Copy chunks to avoid keeping reference to entire chunks array
		batchChunks := make([]string, end-start)
		copy(batchChunks, chunks[start:end])
		fmt.Printf("[批次 %d/%d] 处理块 %d-%d...\n", batchIdx+1, totalBatches, start, end-1)

		// Generate embeddings for this batch
		batchEmbeddings, err := s.generateEmbeddings(ctx, batchChunks)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings for batch %d: %w", batchIdx, err)
		}

		// Store this batch in Qdrant immediately
		err = s.storeInQdrantBatch(ctx, filename, start, batchChunks, batchEmbeddings)
		if err != nil {
			return fmt.Errorf("failed to store batch %d in Qdrant: %w", batchIdx, err)
		}

		// Clear references to help GC
		batchChunks = nil
		batchEmbeddings = nil

		// Force GC every 5 batches to free memory aggressively
		if (batchIdx+1)%5 == 0 {
			runtime.GC()
		}
	}

	fmt.Printf("[完成] 文档已成功处理并存储（共 %d 个块）\n", len(chunks))

	// Clear chunks array to free memory
	chunks = nil

	// Force final GC
	runtime.GC()

	return nil
}

func (s *StorageService) parseDocument(filename string, content []byte) (string, error) {
	// Check file extension
	if strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		return s.parsePDF(content)
	}
	// For other text files, treat as plain text
	return string(content), nil
}

func (s *StorageService) parsePDF(content []byte) (string, error) {
	return parser.ParsePDF(content)
}

// splitText is now implemented in chunker.go with better Chinese support

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"` // Ollama expects string, not array
}

type EmbeddingResponse struct {
	Embedding  []float64   `json:"embedding"`  // Single embedding
	Embeddings [][]float64 `json:"embeddings"` // Multiple embeddings (if supported)
}

// generateEmbeddings generates embeddings for multiple chunks
// Since Ollama API expects a single string, we process chunks sequentially
func (s *StorageService) generateEmbeddings(ctx context.Context, chunks []string) ([][]float64, error) {
	embeddings := make([][]float64, 0, len(chunks))
	// Reuse HTTP client with timeout to avoid connection leaks
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	total := len(chunks)
	for i, chunk := range chunks {
		// Show progress every 10 chunks or on last chunk
		if (i+1)%10 == 0 || i == total-1 {
			fmt.Printf("  [进度] 已处理 %d/%d 个块\n", i+1, total)
		}
		reqBody := EmbeddingRequest{
			Model:  "bge-m3",
			Prompt: chunk,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request for chunk %d: %w", i, err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.EmbeddingURL, strings.NewReader(string(jsonData)))
		if err != nil {
			return nil, fmt.Errorf("failed to create request for chunk %d: %w", i, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to request embedding for chunk %d: %w", i, err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("embedding API error for chunk %d (status %d): %s", i, resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read embedding response for chunk %d: %w", i, err)
		}

		var embeddingResp EmbeddingResponse
		if err := json.Unmarshal(body, &embeddingResp); err != nil {
			// Clear body before returning error to free memory
			body = nil
			return nil, fmt.Errorf("failed to parse embedding response for chunk %d: %w", i, err)
		}

		// Clear body after parsing to free memory immediately
		body = nil

		// Extract embedding (handle both single and multiple formats)
		var embedding []float64
		if len(embeddingResp.Embedding) > 0 {
			embedding = make([]float64, len(embeddingResp.Embedding))
			copy(embedding, embeddingResp.Embedding)
		} else if len(embeddingResp.Embeddings) > 0 {
			embedding = make([]float64, len(embeddingResp.Embeddings[0]))
			copy(embedding, embeddingResp.Embeddings[0])
		} else {
			return nil, fmt.Errorf("no embedding returned for chunk %d", i)
		}

		// Clear embeddingResp to free memory
		embeddingResp = EmbeddingResponse{}

		embeddings = append(embeddings, embedding)
	}

	if len(embeddings) != len(chunks) {
		return nil, fmt.Errorf("mismatch: expected %d embeddings, got %d", len(chunks), len(embeddings))
	}

	return embeddings, nil
}

// generatePointID generates a unique integer ID for a point
// Uses hash of filename + chunk index to ensure uniqueness
func generatePointID(filename string, chunkIdx int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(filename))
	h.Write([]byte(fmt.Sprintf("_%d", chunkIdx)))
	return h.Sum64()
}

// storeInQdrantBatch stores a batch of chunks and embeddings to Qdrant
// This is used for batch processing to avoid memory issues
func (s *StorageService) storeInQdrantBatch(ctx context.Context, filename string, startIdx int, chunks []string, embeddings [][]float64) error {
	// Prepare points for batch upload
	points := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		pointID := generatePointID(filename, startIdx+i)
		points[i] = map[string]interface{}{
			"id":     pointID, // Qdrant requires integer or UUID
			"vector": embeddings[i],
			"payload": map[string]interface{}{
				"text":     chunk,
				"filename": filename,
				"chunk_id": startIdx + i,
			},
		}
	}

	// Upload points to Qdrant
	uploadData := map[string]interface{}{
		"points": points,
	}

	jsonData, err := json.Marshal(uploadData)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/collections/%s/points", s.cfg.QdrantURL, s.cfg.CollectionName)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.QdrantAPIKey != "" {
		req.Header.Set("api-key", s.cfg.QdrantAPIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		// Clear points before returning error
		points = nil
		return fmt.Errorf("Qdrant upload error: %s", string(body))
	}

	// Clear points immediately after successful upload to free memory
	points = nil

	return nil
}

// storeInQdrant stores all chunks and embeddings to Qdrant (kept for backward compatibility)
func (s *StorageService) storeInQdrant(ctx context.Context, filename string, chunks []string, embeddings [][]float64) error {
	return s.storeInQdrantBatch(ctx, filename, 0, chunks, embeddings)
}

func (s *StorageService) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	url := fmt.Sprintf("%s/collections/%s", s.cfg.QdrantURL, s.cfg.CollectionName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	if s.cfg.QdrantAPIKey != "" {
		req.Header.Set("api-key", s.cfg.QdrantAPIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If collection exists, return
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// Create collection if it doesn't exist
	createData := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     1024, // BGE-M3 produces 1024-dimensional vectors
			"distance": "Cosine",
		},
	}

	jsonData, err := json.Marshal(createData)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.QdrantAPIKey != "" {
		req.Header.Set("api-key", s.cfg.QdrantAPIKey)
	}

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: %s", string(body))
	}

	return nil
}
