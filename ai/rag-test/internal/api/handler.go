package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"

	"rag-test/internal/rag"
	"rag-test/internal/storage"
)

type Handler struct {
	ragService     *rag.RAGService
	storageService *storage.StorageService
}

func NewHandler(ragService *rag.RAGService, storageService *storage.StorageService) *Handler {
	return &Handler{
		ragService:     ragService,
		storageService: storageService,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.URL.Path {
	case "/api/upload":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.handleUpload(w, r)
	case "/api/query":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.handleQuery(w, r)
	case "/api/health":
		h.handleHealth(w, r)
	default:
		http.NotFound(w, r)
	}
}

type UploadRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"` // Base64 encoded content
}

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with larger size limit for big files
	// But we'll process in batches to avoid memory issues
	maxMemory := 50 << 20 // 50 MB max memory for form parsing
	err := r.ParseMultipartForm(int64(maxMemory))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Printf("[上传] 收到文件: %s\n", header.Filename)
	
	// Read file content - for very large files, this could still be an issue
	// But PDF parsing requires the full content, so we need to read it
	// The batch processing in UploadDocument will help with memory after parsing
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileSize := len(content)
	fmt.Printf("[上传] 文件大小: %d 字节 (%.2f MB)\n", fileSize, float64(fileSize)/(1024*1024))
	
	// Clear the file reference to help GC
	file = nil

	err = h.storageService.UploadDocument(r.Context(), header.Filename, content)
	if err != nil {
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Note: content is already cleared in UploadDocument
	// Force GC after processing
	runtime.GC()

	json.NewEncoder(w).Encode(UploadResponse{
		Success: true,
		Message: "Document uploaded and processed successfully",
	})
}

type QueryRequest struct {
	Question string `json:"question"`
}

func (h *Handler) handleQuery(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	response, err := h.ragService.Query(r.Context(), req.Question)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
