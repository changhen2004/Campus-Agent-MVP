package qdrant

import (
	"context"
	"fmt"
	"log"

	"campus-agent/pkg/config"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn        *grpc.ClientConn
	points      pb.PointsClient
	collections pb.CollectionsClient
	collection  string
	dimension   uint64
}

func NewClient(ctx context.Context, cfg config.QdrantConfig, dimension int) (*Client, error) {
	addr := cfg.Addr()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("qdrant dial %s: %w", addr, err)
	}

	pointsClient := pb.NewPointsClient(conn)
	collectionsClient := pb.NewCollectionsClient(conn)

	client := &Client{
		conn:        conn,
		points:      pointsClient,
		collections: collectionsClient,
		collection:  cfg.Collection,
		dimension:   uint64(dimension),
	}

	if err := client.ensureCollection(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("qdrant ensure collection: %w", err)
	}

	log.Printf("[qdrant] connected to %s, collection=%s, dimension=%d", addr, cfg.Collection, dimension)
	return client, nil
}

func (c *Client) ensureCollection(ctx context.Context) error {
	// 检查 collection 是否存在
	_, err := c.collections.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: c.collection,
	})
	if err == nil {
		return nil // 已存在
	}

	// 创建 collection
	_, err = c.collections.Create(ctx, &pb.CreateCollection{
		CollectionName: c.collection,
		VectorsConfig: pb.NewVectorsConfig(&pb.VectorParams{
			Size:     c.dimension,
			Distance: pb.Distance_Cosine,
		}),
	})
	if err != nil {
		// Collection 可能已被其他进程创建
		log.Printf("[qdrant] create collection: %v (may already exist)", err)
	}

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CollectionName() string {
	return c.collection
}
