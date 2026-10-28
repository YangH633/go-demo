package parser

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ParsePDF extracts text from PDF file using pdftotext
// This requires pdftotext to be installed (part of poppler package)
func ParsePDF(content []byte) (string, error) {
	fmt.Printf("[PDF解析] 开始解析PDF，文件大小: %d 字节\n", len(content))

	// Check if pdftotext is available
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", fmt.Errorf("pdftotext not found. Please install it:\n  macOS: brew install poppler\n  Ubuntu/Debian: sudo apt-get install poppler-utils\n\nAfter installation, restart the server.")
	}

	// Write content to temporary file
	fmt.Printf("[PDF解析] 创建临时文件...\n")
	tmpFile, err := os.CreateTemp("", "pdf_*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(content); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()
	fmt.Printf("[PDF解析] 临时文件已写入\n")

	// Use pdftotext to extract text with timeout
	fmt.Printf("[PDF解析] 调用 pdftotext 提取文本（这可能需要一些时间）...\n")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", tmpFile.Name(), "-")

	startTime := time.Now()
	output, err := cmd.Output()
	duration := time.Since(startTime)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("PDF解析超时（超过5分钟），文件可能太大或太复杂")
		}
		return "", fmt.Errorf("failed to extract PDF text: %w", err)
	}

	fmt.Printf("[PDF解析] pdftotext 完成，耗时: %.2f 秒\n", duration.Seconds())

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("no text extracted from PDF (PDF might be image-based or encrypted)")
	}

	fmt.Printf("[PDF解析] 提取文本成功，文本长度: %d 字符\n", len(result))
	return result, nil
}
