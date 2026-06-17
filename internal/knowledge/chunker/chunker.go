package chunker

import (
	"strings"
	"unicode/utf8"
)

// Chunk 表示单个文本片段及其位置元数据。
type Chunk struct {
	Content string
	Index   int // 在文档中的 0 基位置
}

// 针对中文文本优化的默认值，适配约 1500–2000 token 的 embedding 模型。
const (
	defaultChunkSize    = 2000 // 字符数
	defaultChunkOverlap = 200  // 字符数
)

// Split 使用递归分割将文本拆分为有重叠的块：
//
//	段落（双换行）→ 句子（。！？）→ 固定大小分割
//
// 每个块 ≤ chunkSize 个字符，相邻块之间有重叠。
func Split(text string) []Chunk {
	return SplitWithSize(text, defaultChunkSize, defaultChunkOverlap)
}

// SplitWithSize 与 Split 类似，但接受自定义块大小和重叠量。
func SplitWithSize(text string, chunkSize, overlap int) []Chunk {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if utf8.RuneCountInString(text) <= chunkSize {
		return []Chunk{{Content: text, Index: 0}}
	}

	// 第一阶段：按双换行（段落）分割
	paragraphs := splitParagraphs(text)
	// 第二阶段：将长段落拆分为句子
	sentences := splitLongSegments(paragraphs, chunkSize, sentSplitter)
	// 第三阶段：将长句子拆分为固定大小的片段
	segments := splitLongSegments(sentences, chunkSize, fixedSplitter)

	// 将片段合并为有重叠的块
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
	// 按中英文句子边界分割
	var parts []string
	start := 0
	runes := []rune(text)

	for i, r := range runes {
		if r == '。' || r == '！' || r == '？' || r == '\n' || r == '.' || r == '!' || r == '?' {
			if i+1 < len(runes) && runes[i+1] == ' ' {
				// 处理 ". " 等格式
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
			// 以重叠方式开始新块：保留上一个块末尾的若干片段
			current = nil
			currentLen = 0

			// 从前一个块的尾部提取重叠前缀
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
