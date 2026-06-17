package qdrant

import (
	"context"
	"crypto/md5"
	"fmt"

	"campus-agent/internal/ai/embedder"
	"campus-agent/internal/knowledge/chunker"

	pb "github.com/qdrant/go-client/qdrant"
)

type Indexer struct {
	client   *Client
	embedder *embedder.Embedder
}

func NewIndexer(client *Client, embedder *embedder.Embedder) *Indexer {
	return &Indexer{client: client, embedder: embedder}
}

// Document 表示一条待索引的知识文档。
type Document struct {
	ID      string
	Title   string
	Content string
}

// IndexDocuments 将每个文档拆分为块，嵌入向量后 upsert 到 Qdrant。
// 超过 embedding API 请求大小限制的长文档会自动切割并保留重叠以维持跨块上下文。
func (idx *Indexer) IndexDocuments(ctx context.Context, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	points := make([]*pb.PointStruct, 0, len(docs)*2)
	for _, doc := range docs {
		text := doc.Title
		if doc.Content != "" {
			text = doc.Title + "\n" + doc.Content
		}

		chunks := chunker.Split(text)
		if len(chunks) == 0 {
			chunks = []chunker.Chunk{{Content: text, Index: 0}}
		}

		for _, chunk := range chunks {
			vector64, err := idx.embedder.Embed(ctx, chunk.Content)
			if err != nil {
				return fmt.Errorf("embed doc %s chunk %d: %w", doc.ID, chunk.Index, err)
			}

			vector32 := make([]float32, len(vector64))
			for i, v := range vector64 {
				vector32[i] = float32(v)
			}

			chunkID := doc.ID
			if len(chunks) > 1 {
				chunkID = fmt.Sprintf("%s#%d", doc.ID, chunk.Index)
			}

			points = append(points, &pb.PointStruct{
				Id:      pb.NewIDUUID(toUUID(chunkID)),
				Vectors: pb.NewVectors(vector32...),
				Payload: map[string]*pb.Value{
					"title":       {Kind: &pb.Value_StringValue{StringValue: doc.Title}},
					"content":     {Kind: &pb.Value_StringValue{StringValue: chunk.Content}},
					"doc_id":      {Kind: &pb.Value_StringValue{StringValue: doc.ID}},
					"chunk_index": {Kind: &pb.Value_IntegerValue{IntegerValue: int64(chunk.Index)}},
				},
			})

			// 每 64 个点批量 upsert
			if len(points) >= 64 {
				if err := idx.upsertBatch(ctx, points); err != nil {
					return err
				}
				points = points[:0]
			}
		}
	}

	if len(points) > 0 {
		return idx.upsertBatch(ctx, points)
	}
	return nil
}

// toUUID 使用 MD5 哈希将任意字符串转换为合法的 UUID v3/v5 格式。
// Qdrant 要求 point ID 为 UUID，但我们的文档 ID 是可读字符串（如中文文件名）。
// 此哈希产生确定性 UUID，确保同一文档重新索引时始终生成相同的 Qdrant point ID。
func toUUID(s string) string {
	hash := md5.Sum([]byte(s))
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
}

func (idx *Indexer) upsertBatch(ctx context.Context, points []*pb.PointStruct) error {
	batch := make([]*pb.PointStruct, len(points))
	copy(batch, points)
	_, err := idx.client.points.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: idx.client.collection,
		Points:         batch,
	})
	if err != nil {
		return fmt.Errorf("upsert batch: %w", err)
	}
	return nil
}
