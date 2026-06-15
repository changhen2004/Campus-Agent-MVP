package qdrant

import (
	"context"
	"fmt"

	"campus-agent/internal/ai/embedder"

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

// IndexDocuments embeds and upserts documents into Qdrant.
func (idx *Indexer) IndexDocuments(ctx context.Context, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	points := make([]*pb.PointStruct, 0, len(docs))
	for i, doc := range docs {
		text := doc.Title
		if doc.Content != "" {
			text = doc.Title + "\n" + doc.Content
		}

		vector64, err := idx.embedder.Embed(ctx, text)
		if err != nil {
			return fmt.Errorf("embed doc %s: %w", doc.ID, err)
		}

		// Convert float64 to float32 for Qdrant
		vector32 := make([]float32, len(vector64))
		for j, v := range vector64 {
			vector32[j] = float32(v)
		}

		points = append(points, &pb.PointStruct{
			Id:      pb.NewIDUUID(doc.ID),
			Vectors: pb.NewVectors(vector32...),
			Payload: map[string]*pb.Value{
				"title":   {Kind: &pb.Value_StringValue{StringValue: doc.Title}},
				"content": {Kind: &pb.Value_StringValue{StringValue: doc.Content}},
			},
		})

		// Batch upsert every 64 documents
		if len(points) >= 64 || i == len(docs)-1 {
			batch := make([]*pb.PointStruct, len(points))
			copy(batch, points)
			points = points[:0]

			_, err := idx.client.points.Upsert(ctx, &pb.UpsertPoints{
				CollectionName: idx.client.collection,
				Points:         batch,
			})
			if err != nil {
				return fmt.Errorf("upsert batch: %w", err)
			}
		}
	}

	return nil
}
