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

// Document represents a knowledge document to be indexed.
type Document struct {
	ID      string
	Title   string
	Content string
}

// IndexDocuments splits each document into chunks, embeds them, and upserts into Qdrant.
// Long documents that exceed the embedding API's request size limit are automatically
// split with overlap to preserve cross-chunk context.
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

			// Batch upsert every 64 points
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

// toUUID converts an arbitrary string to a valid UUID v3/v5 format using MD5 hash.
// Qdrant requires UUIDs for point IDs, but our document IDs are human-readable strings
// (e.g. Chinese filenames). This hash produces a deterministic UUID so re-indexing
// the same document always produces the same Qdrant point ID.
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
