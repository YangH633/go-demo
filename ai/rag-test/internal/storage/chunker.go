package storage

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// splitText intelligently splits text into chunks, with better support for Chinese
func (s *StorageService) splitText(text string) []string {
	fmt.Printf("[分块] 开始清理文本...\n")
	// Clean text: remove excessive whitespace
	text = cleanText(text)
	fmt.Printf("[分块] 文本清理完成，开始分块处理...\n")

	chunks := []string{}
	runes := []rune(text)
	fmt.Printf("[分块] 文本转换为rune数组，长度: %d\n", len(runes))

	chunkSize := s.cfg.ChunkSize
	overlap := s.cfg.ChunkOverlap
	fmt.Printf("[分块] 分块参数: chunkSize=%d, overlap=%d\n", chunkSize, overlap)

	// Chinese sentence delimiters: 。！？；\n
	chineseSentenceEnd := regexp.MustCompile(`[。！？；\n]+`)
	fmt.Printf("[分块] 开始循环分块...\n")

	i := 0
	chunkCount := 0
	maxIterations := len(runes) * 2 // Safety limit
	iterationCount := 0

	for i < len(runes) {
		iterationCount++
		if iterationCount > maxIterations {
			fmt.Printf("[分块] 错误: 超过最大迭代次数，可能存在死循环\n")
			break
		}

		chunkCount++
		if chunkCount%10 == 0 {
			fmt.Printf("[分块] 已处理到位置 %d/%d (迭代 %d)\n", i, len(runes), iterationCount)
		}
		// Try to find a good break point near the target chunk size
		targetEnd := i + chunkSize
		if targetEnd > len(runes) {
			targetEnd = len(runes)
		}

		// Look for sentence boundaries within a reasonable range
		searchStart := targetEnd - 100 // Look back 100 characters
		if searchStart < i {
			searchStart = i
		}
		searchEnd := targetEnd + 50 // Look forward 50 characters
		if searchEnd > len(runes) {
			searchEnd = len(runes)
		}

		// Find the best break point (sentence boundary)
		bestBreak := targetEnd
		searchText := string(runes[searchStart:searchEnd])
		matches := chineseSentenceEnd.FindAllStringIndex(searchText, -1)

		if len(matches) > 0 {
			// Find the match closest to targetEnd
			for _, match := range matches {
				pos := searchStart + match[1]
				if pos >= i && pos <= targetEnd+50 {
					if abs(pos-targetEnd) < abs(bestBreak-targetEnd) {
						bestBreak = pos
					}
				}
			}
		}

		// If no good break point found, use targetEnd
		if bestBreak <= i {
			bestBreak = targetEnd
		}

		// Ensure bestBreak is at least i+1 to make progress
		if bestBreak <= i {
			bestBreak = i + 1
		}

		// If we've reached the end, take the rest
		if bestBreak >= len(runes) {
			bestBreak = len(runes)
		}

		chunk := string(runes[i:bestBreak])
		chunk = strings.TrimSpace(chunk)
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}

		// Move to next chunk with overlap
		// Ensure we always make progress (at least 1 character forward)
		nextI := bestBreak - overlap
		if nextI <= i {
			// Force progress: move forward at least 1 character
			nextI = i + 1
		}
		i = nextI

		if i >= len(runes) {
			break
		}
	}

	fmt.Printf("[分块] 分块完成，共生成 %d 个块\n", len(chunks))
	return chunks
}

// cleanText removes excessive whitespace and normalizes text
func cleanText(text string) string {
	// Remove excessive newlines (more than 2 consecutive)
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	// Remove excessive spaces (more than 2 consecutive)
	text = regexp.MustCompile(` {3,}`).ReplaceAllString(text, "  ")

	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

// abs returns absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// isChinese checks if a rune is a Chinese character
func isChinese(r rune) bool {
	return unicode.Is(unicode.Han, r)
}
