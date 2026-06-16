package chunker

import (
	"strings"
	"unicode/utf8"
)

// Chunk represents a single text fragment with position metadata.
type Chunk struct {
	Content string
	Index   int // 0-based position within the document
}

// Defaults tuned for Chinese text with ~1500-2000 token embedding models.
const (
	defaultChunkSize    = 2000 // characters
	defaultChunkOverlap = 200  // characters
)

// Split breaks text into overlapping chunks using recursive splitting:
//
//	paragraphs (double newline) → sentences (。！？) → fixed-size
//
// Each chunk is ≤ chunkSize characters, with overlap between adjacent chunks.
func Split(text string) []Chunk {
	return SplitWithSize(text, defaultChunkSize, defaultChunkOverlap)
}

// SplitWithSize is like Split but accepts custom chunk size and overlap.
func SplitWithSize(text string, chunkSize, overlap int) []Chunk {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if utf8.RuneCountInString(text) <= chunkSize {
		return []Chunk{{Content: text, Index: 0}}
	}

	// Stage 1: split on double newline (paragraphs)
	paragraphs := splitParagraphs(text)
	// Stage 2: split long paragraphs into sentences
	sentences := splitLongSegments(paragraphs, chunkSize, sentSplitter)
	// Stage 3: split long sentences into fixed-size pieces
	segments := splitLongSegments(sentences, chunkSize, fixedSplitter)

	// Merge segments into overlapping chunks
	return mergeChunks(segments, chunkSize, overlap)
}

func splitParagraphs(text string) []string {
	parts := strings.Split(text, "\n\n")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

type splitFunc func(string, int) []string

func sentSplitter(text string, _ int) []string {
	// Split on Chinese/English sentence boundaries
	var parts []string
	start := 0
	runes := []rune(text)

	for i, r := range runes {
		if r == '。' || r == '！' || r == '？' || r == '\n' || r == '.' || r == '!' || r == '?' {
			if i+1 < len(runes) && runes[i+1] == ' ' {
				// handle ". " etc.
			}
			if i > start {
				parts = append(parts, string(runes[start:i+1]))
			}
			start = i + 1
		}
	}
	if start < len(runes) {
		parts = append(parts, string(runes[start:]))
	}
	return parts
}

func fixedSplitter(text string, limit int) []string {
	runes := []rune(text)
	if len(runes) <= limit {
		return []string{text}
	}

	var parts []string
	for i := 0; i < len(runes); i += limit {
		end := i + limit
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[i:end]))
	}
	return parts
}

func splitLongSegments(segments []string, limit int, fn splitFunc) []string {
	var result []string
	for _, seg := range segments {
		if utf8.RuneCountInString(seg) <= limit {
			if strings.TrimSpace(seg) != "" {
				result = append(result, seg)
			}
			continue
		}
		parts := fn(seg, limit)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
	}
	return result
}

func mergeChunks(segments []string, chunkSize, overlap int) []Chunk {
	if len(segments) == 0 {
		return nil
	}

	var chunks []Chunk
	var current []string
	currentLen := 0
	index := 0

	flush := func() {
		content := strings.Join(current, "")
		if strings.TrimSpace(content) != "" {
			chunks = append(chunks, Chunk{Content: content, Index: index})
			index++
		}
	}

	for _, seg := range segments {
		segLen := utf8.RuneCountInString(seg)

		if currentLen+segLen > chunkSize && len(current) > 0 {
			flush()
			// Start new chunk with overlap: carry over last few segments
			current = nil
			currentLen = 0

			// Build overlap prefix from the tail of the previous chunk
			overlapContent := overlapFromLast(chunks, overlap)
			if overlapContent != "" {
				current = append(current, overlapContent)
				currentLen = utf8.RuneCountInString(overlapContent)
			}
		}

		current = append(current, seg)
		currentLen += segLen
	}

	flush()
	return chunks
}

func overlapFromLast(chunks []Chunk, targetOverlap int) string {
	if len(chunks) == 0 {
		return ""
	}

	last := chunks[len(chunks)-1].Content
	runes := []rune(last)
	if len(runes) <= targetOverlap {
		return string(runes)
	}
	return string(runes[len(runes)-targetOverlap:])
}
