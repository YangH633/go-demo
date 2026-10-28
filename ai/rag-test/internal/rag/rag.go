package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rag-test/internal/config"
)

type RAGService struct {
	cfg *config.Config
}

func NewRAGService(cfg *config.Config) *RAGService {
	return &RAGService{cfg: cfg}
}

type SearchResult struct {
	Text     string  `json:"text"`
	Filename string  `json:"filename"`
	Score    float64 `json:"score"`
}

type QueryResponse struct {
	Answer  string         `json:"answer"`
	Sources []SearchResult `json:"sources"`
}

// Query performs RAG: retrieves relevant documents and generates answer
func (r *RAGService) Query(ctx context.Context, question string) (*QueryResponse, error) {
	// Step 1: Generate embedding for the question
	questionEmbedding, err := r.generateQuestionEmbedding(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	// Step 2: Search in Qdrant
	results, err := r.searchInQdrant(ctx, questionEmbedding, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search in Qdrant: %w", err)
	}

	// Step 3: Build context from retrieved documents
	context := r.buildContext(results)

	// Step 4: Generate answer using Ollama
	answer, err := r.generateAnswer(ctx, question, context)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	return &QueryResponse{
		Answer:  answer,
		Sources: results,
	}, nil
}

func (r *RAGService) generateQuestionEmbedding(ctx context.Context, question string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model":  "bge-m3",
		"prompt": question, // Ollama expects string, not array
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.cfg.EmbeddingURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}

	var embeddingResp struct {
		Embedding  []float64   `json:"embedding"`  // Single embedding
		Embeddings [][]float64 `json:"embeddings"` // Multiple embeddings
	}
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w, body: %s", err, string(body))
	}

	// Handle both single and multiple embedding formats
	if len(embeddingResp.Embedding) > 0 {
		return embeddingResp.Embedding, nil
	}
	if len(embeddingResp.Embeddings) > 0 {
		return embeddingResp.Embeddings[0], nil
	}

	return nil, fmt.Errorf("no embedding returned in response")
}

func (r *RAGService) searchInQdrant(ctx context.Context, vector []float64, limit int) ([]SearchResult, error) {
	searchData := map[string]interface{}{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
	}

	jsonData, err := json.Marshal(searchData)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", r.cfg.QdrantURL, r.cfg.CollectionName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if r.cfg.QdrantAPIKey != "" {
		req.Header.Set("api-key", r.cfg.QdrantAPIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Qdrant search error: %s", string(body))
	}

	var searchResp struct {
		Result []struct {
			Score   float64 `json:"score"`
			Payload struct {
				Text     string `json:"text"`
				Filename string `json:"filename"`
			} `json:"payload"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(searchResp.Result))
	for i, item := range searchResp.Result {
		results[i] = SearchResult{
			Text:     item.Payload.Text,
			Filename: item.Payload.Filename,
			Score:    item.Score,
		}
	}

	return results, nil
}

func (r *RAGService) buildContext(results []SearchResult) string {
	var builder strings.Builder
	builder.WriteString("相关文档内容：\n\n")
	for i, result := range results {
		builder.WriteString(fmt.Sprintf("文档片段 %d (来源: %s, 相似度: %.3f):\n", i+1, result.Filename, result.Score))
		builder.WriteString(result.Text)
		builder.WriteString("\n\n")
	}
	return builder.String()
}

func (r *RAGService) generateAnswer(ctx context.Context, question, context string) (string, error) {
	prompt := fmt.Sprintf(`基于以下文档内容回答问题。如果文档中没有相关信息，请说明。

%s

问题：%s

请用中文回答：`, context, question)

	reqBody := map[string]interface{}{
		"model":  "qwen2.5:7b",
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.cfg.OllamaURL+"/api/generate", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var generateResp struct {
		Response string `json:"response"`
	}

	if err := json.Unmarshal(body, &generateResp); err != nil {
		// Try to extract response from stream format if it's not a single JSON object
		// Ollama might return multiple JSON objects (streaming format)
		lines := strings.Split(string(body), "\n")
		var response strings.Builder
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var chunk struct {
				Response string `json:"response"`
			}
			if err := json.Unmarshal([]byte(line), &chunk); err == nil {
				response.WriteString(chunk.Response)
			}
		}
		if response.Len() > 0 {
			return response.String(), nil
		}
		return "", fmt.Errorf("failed to parse Ollama response: %s", string(body))
	}

	return generateResp.Response, nil
}
